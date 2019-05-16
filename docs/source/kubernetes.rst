
Kubernetes Extensions
=====================

Official Documentation
----------------------

`Kubernetes Extension <https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/>`_

Notes
-----

1. Definitions of a CRD is as flexible as the definition of Helm values.yaml.
   This lake of consistency accross helm values.yaml and CRD definitions,
   brings the same level of complexity at integration time.
2. Controller are permanently receiving events for the CRD is watches:

   - No need to poll the service or the CRD and subresources.
   - There are exceptions: Pod blocked in the "initcontainer" phase, will never send anything. No state transition. How do
     we ensure proper timeout handling.


