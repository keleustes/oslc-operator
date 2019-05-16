
ETCD lifecycle
==============

Installation Phase
------------------

1. `OSHM ETCD Chart <https://github.com/openstack/openstack-helm-infra/tree/master/etcd>`_
2. `Promenade ETCD Chart <https://github.com/openstack/airship-promenade/tree/master/charts/etcd>`_


Operational Phase
-----------------

1. `CronJob <https://github.com/openstack/airship-promenade/blob/master/charts/etcd/templates/cron-job-etcd-backup.yaml>`_
2. `ShellScript <https://github.com/openstack/airship-promenade/blob/master/charts/etcd/templates/bin/_etcdbackup.tpl>`_

Backup
------

TBD

Restore
-------

TBD

Upgrade
-------

TBD


CEPH lifecycle
==============

Installation Phase
------------------

1. `ceph client <https://github.com/openstack/openstack-helm-infra/tree/master/ceph-client>`_
2. `ceph mon <https://github.com/openstack/openstack-helm-infra/tree/master/ceph-mon>`_
3. `ceph osd <https://github.com/openstack/openstack-helm-infra/tree/master/ceph-osd>`_
4. `ceph provisioners <https://github.com/openstack/openstack-helm-infra/tree/master/ceph-provisioners>`_
5. `cep rgw <https://github.com/openstack/openstack-helm-infra/tree/master/ceph-rgw>`_


Backup
------

1. `Ceph utility Container <https://github.com/att-comdev/porthole/tree/master/ceph-utility>`_

Restore
-------

TBD

Upgrade
-------

TBD

Calico lifecycle
================

Installation
------------

1. `OSHM Calico Chart <https://github.com/openstack/openstack-helm-infra/tree/master/calico>`_

Backup
------

1. `Calico utility Container <https://github.com/att-comdev/porthole/tree/master/calicoctl-utility>`_

Restore
-------

TBD

Upgrade
-------

TBD


MariaDB lifecycle
=================

Backup
------

TBD

Restore
-------

TBD

Upgrade
-------

TBD

