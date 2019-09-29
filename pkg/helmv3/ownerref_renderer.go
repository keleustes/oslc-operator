// Copyright The Helm Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build v3

package helmv3

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/releaseutil"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/timeconv"

	av1 "github.com/keleustes/armada-crd/pkg/apis/openstacklcm/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	log = logf.Log.WithName("helmv3")

	// defaultKubeVersion is the default value of --kube-version flag
	defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)
)

type OwnerRefHelmv3Renderer struct {
	refs         []metav1.OwnerReference
	suffix       string
	renderFiles  []string
	renderValues map[string]interface{}
}

// Adds the ownerrefs to all the documents in a YAML file
func (o *OwnerRefHelmv3Renderer) RenderFile(name string, namespace string, fileName string) (*av1.SubResourceList, error) {

	yamlfmt, ferr := ioutil.ReadFile(fileName)
	if ferr != nil {
		log.Error(ferr, "Can not read file")
		return av1.NewSubResourceList(namespace, name), ferr
	}
	ownedRenderedFiles, err := o.fromYaml(name, namespace, string(yamlfmt))
	if err != nil {
		log.Info("Can not convert malformed yaml to unstructured", "filename", fileName)
		return ownedRenderedFiles, err
	}

	return ownedRenderedFiles, nil
}

// Renders an entire chart and adds the ownerref
func (o *OwnerRefHelmv3Renderer) RenderChart(name string, namespace string, chartLocation string) (*av1.SubResourceList, error) {

	ownedRenderedFiles := av1.NewSubResourceList(namespace, name)

	// verify chart path exists
	var chartPath string
	if _, err := os.Stat(chartLocation); err == nil {
		if chartPath, err = filepath.Abs(chartLocation); err != nil {
			return ownedRenderedFiles, err
		}
	} else {
		return ownedRenderedFiles, err
	}

	// get combined values and create config
	rawVals, err := yaml.Marshal(o.renderValues)
	if err != nil {
		return ownedRenderedFiles, err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.Load(chartPath)
	if err != nil {
		return ownedRenderedFiles, err
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      name,
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: namespace,
		},
		KubeVersion: defaultKubeVersion,
	}

	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return ownedRenderedFiles, err
	}

	listManifests := manifest.SplitManifests(renderedTemplates)
	var manifestsToRender []manifest.Manifest

	// if we have a list of files to render, then check that each of the
	// provided files exists in the chart.
	if len(o.renderFiles) > 0 {
		for _, f := range o.renderFiles {
			missing := true
			if !filepath.IsAbs(f) {
				newF, err := filepath.Abs(filepath.Join(chartPath, f))
				if err != nil {
					return ownedRenderedFiles, fmt.Errorf("could not turn template path %s into absolute path: %s", f, err)
				}
				f = newF
			}

			for _, manifest := range listManifests {
				// manifest.Name is rendered using linux-style filepath separators on Windows as
				// well as macOS/linux.
				manifestPathSplit := strings.Split(manifest.Name, "/")
				// remove the chart name from the path
				manifestPathSplit = manifestPathSplit[1:]
				toJoin := append([]string{chartPath}, manifestPathSplit...)
				manifestPath := filepath.Join(toJoin...)

				// if the filepath provided matches a manifest path in the
				// chart, render that manifest
				if f == manifestPath {
					manifestsToRender = append(manifestsToRender, manifest)
					missing = false
				}
			}
			if missing {
				return ownedRenderedFiles, fmt.Errorf("could not find template %s in chart", f)
			}
		}
	} else {
		// no renderFiles provided, render all manifests in the chart
		manifestsToRender = listManifests
	}

	for _, m := range manifestsToRender {
		fileName := filepath.Base(m.Name)
		if fileName == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(fileName, "_") {
			continue
		}

		// TODO(jeb): We need to figure how to reuse configmap
		// if strings.HasSuffix(fileName, ".yaml") && strings.Contains(m.Name, "templates/lifecycle-") {
		if strings.HasSuffix(fileName, ".yaml") {
			parsed, _ := o.fromYaml(name, namespace, m.Content)

			for _, u := range parsed.Items {
				ownedRenderedFiles.Items = append(ownedRenderedFiles.Items, u)
			}
		}
	}

	return ownedRenderedFiles, nil
}

func sortManifests(in map[string]string) []string {
	var keys []string
	for k := range in {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var manifests []string
	for _, k := range keys {
		manifests = append(manifests, in[k])
	}
	return manifests
}

// Reads a yaml file and converts into an Unstructured object
func (o *OwnerRefHelmv3Renderer) fromYaml(name string, namespace string, filecontent string) (*av1.SubResourceList, error) {

	ownedRenderedFiles := av1.NewSubResourceList(namespace, name)

	manifests := releaseutil.SplitManifests(filecontent)
	for _, manifest := range sortManifests(manifests) {
		manifestMap := chartutil.FromYaml(manifest)

		if _, ok := manifestMap["Error"]; ok {
			log.Error(nil, "error parsing rendered template to add ownerrefs")
			continue
		}

		// Check if the document is empty
		if len(manifestMap) == 0 {
			continue
		}

		unst, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&manifestMap)
		if err != nil {
			log.Error(err, "error converting to Unstructured")
			continue
		}

		u := &unstructured.Unstructured{Object: unst}
		u.SetOwnerReferences(o.refs)

		// Init name and namespace
		if u.GetName() == "" {
			u.SetName(name + "-" + o.suffix)
		}

		if u.GetNamespace() == "" {
			u.SetNamespace(namespace)
		}

		// Add OwnerReferences
		u.SetOwnerReferences(o.refs)

		// Add labels
		// labels := map[string]string{
		// 	"app": name,
		// }
		// u.SetLabels(labels)

		ownedRenderedFiles.Items = append(ownedRenderedFiles.Items, *u)
	}

	return ownedRenderedFiles, nil
}

// NewOwnerRefHelmv3Renderer creates a new OwnerRef engine with a set of metav1.OwnerReferences to be added to assets
func NewOwnerRefHelmv3Renderer(refs []metav1.OwnerReference, suffix string,
	renderFiles []string, renderValues map[string]interface{}) *OwnerRefHelmv3Renderer {
	return &OwnerRefHelmv3Renderer{
		refs:         refs,
		suffix:       suffix,
		renderFiles:  renderFiles,
		renderValues: renderValues,
	}
}
