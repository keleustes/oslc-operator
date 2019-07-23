module github.com/keleustes/oslc-operator

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v0.0.0-20190301161902-9f8fceff796f // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/google/uuid v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/onsi/ginkgo v1.7.0 // indirect
	github.com/onsi/gomega v1.4.3 // indirect
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829 // indirect
	github.com/prometheus/common v0.4.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190522155817-f3200d17e092 // indirect
	golang.org/x/sys v0.0.0-20190515120540-06a5c4944438 // indirect
	golang.org/x/text v0.3.1 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0 // indirect
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/helm v2.14.1+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.3
)

replace k8s.io/api => k8s.io/api v0.0.0-20190718183219-b59d8169aab5

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190718185103-d1ef975d28ce

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190612205821-1799e75a0719

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190718184206-a1aa83af71a7

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190718185405-0ce9869d0015

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190718183610-8e956561bbf5

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190718190308-f8e43aa19282

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190718190146-f7b0473036f9

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20190718183727-0ececfbe9772

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190531030430-6117653b35f1

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190718190424-bef8d46b95de

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190718184434-a064d4d1ed7a

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190718190030-ea930fedc880

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190718185641-5233cb7cb41e

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190718185913-d5429d807831

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190718185757-9b45f80d5747

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190718190548-039b99e58dbd

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20190718185242-1e1642704fe6

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190718184639-baafa86838c0
