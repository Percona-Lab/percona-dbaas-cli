List of Options
----------------------------------

Any combination of valid options from the Percona XtraDB Cluster Operator `Custom Resource <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html>`_ can be set as a comma-separated list using the ``--options`` flag:

.. code-block:: bash

   percona-dbaas mysql create-db cluster1 --options="proxysql.serviceType=LoadBalancer,proxysql.size=3,pxc.volumeSpec.persistentVolumeClaim.resources.requests=storage:250Gi"

Here is the list of valid commonly used options:

* `pxc.size <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#pxc-size>`_ - The number of members of your PXC database cluster.  Must be an odd number (1, 3, 5, 7, et al)
* `pxc.affinity <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#pxc-affinity-topologykey>`_ - Sets an affinity rule, such as for multi-AZ deployments
* `pxc.resources <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#pxc-resources-requests-memory>`_ - Sets CPU and Memory resources allocated to each member of the PXC database cluster.
* `proxysql.size <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#proxysql-size>`_ - The number of members of your ProxySQL cluster. Recommended to be 1 or 3.
* `proxysql.affinity <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#proxysql-affinity-topologykey>`_ - Sets an affinity rule, such as for multi-AZ deployments
* `proxysql.resources <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#proxysql-resources-requests-memory>`_ - Sets CPU and Memory resources allocated to each member of the ProxySQL cluster.
* `proxysql.enabled <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html#proxysql-enabled>`_ - Defaults to true, allows you to disable proxying.
* `allowUnsafeConfigurations <operator.html#operator-custom-resource-options>`_ - Enables or disables cluster configurations with unsafe parameters, e.g. with less than 3 nodes or without TLS/SSL certificates.
