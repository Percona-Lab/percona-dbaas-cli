List of Options
----------------------------------

Any valid option from the Percona XtraDB Cluster Operator `Custom Resource <https://www.percona.com/doc/kubernetes-operator-for-pxc/operator.html>`_ can be set using the ``--options`` flag.

Here is the list of valid commonly used options:

* ``pxc.size`` - The number of members of your PXC database cluster.  Must be
  an odd number (1, 3, 5, 7, et al)
* ``pxc.affinity`` - Sets an affinity rule, such as for multi-AZ deployments
* ``pxc.resources`` - Sets CPU and Memory resources allocated to each member of
  the PXC database cluster.
* ``proxysql.size`` - The number of members of your ProxySQL cluster.
  Recommended to be 1 or 3.
* ``proxysql.affinity`` - Sets an affinity rule, such as for multi-AZ
  deployments
* ``proxysql.resources`` - Sets CPU and Memory resources allocated to each
  member of the ProxySQL cluster.
* ``proxysql.enabled`` - Defaults to true, allows you to disable proxying.


