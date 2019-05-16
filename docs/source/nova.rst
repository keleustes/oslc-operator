
Nova Deployment
===========================

Schema
----------------------

.. image:: diagrams/nova_greenfieldflow.png

Rationale
---------

1. Ops Team need to deploy a new service.
2. If the service is unhealty, it gets removed.
3. If the service is healty, it reaches the operational.

Nova Change
==================

Schema
------

.. image:: diagrams/nova_brownfieldflow.png

Rationale
---------

1. Ops Team need to:

   - Use Case 1: remove a service.
   - Use Case 2: update a service.
   - Use Case 3: rollback a service.

2. Once the traffic is drain:

   - Use Case 1: the service is removed.
   - Use Case 2: the service is updated.
   - Use Case 3: the service is rollback.

3. Once the update/rollback is performed, the traffic is rollout.
