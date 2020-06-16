List of Options
----------------------------------

Any valid option from the Percona Server for MongoDB Operator `Custom Resource <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html>`_ can be set using the ``--options`` flag:

.. code-block:: bash

   percona-dbaas mongodb create-db cluster1 --options="replsets.arbiter.enabled=true,replsets.size=4,replsets.volumeSpec.persistentVolumeClaim.resources.requests=storage:250Gi"

Here is the list of the valid commonly used options:

* `replsets.size <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-size>`_ - The number of members in the ReplSet. Can be any valid number, although we recommend odd numbers less than 15.
* `replsets.poddisruptionbudget.maxunavailable <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-poddisruptionbudget-maxunavailable>`_ - The amount of members in the ReplSet which can become unavailable within Kubernetes constraints.
* `replsets.affinity <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-affinity-antiaffinitytopologykey>`_ - Lets you set affinity rules.  For instance, you can use this for Multi-AZ deployments.
* `replsets.arbiter.enabled <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-arbiter-enabled>`_ - Lets you turn on the MongoDB arbiter for 2-member ReplSets
* `replsets.resources <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-resources-limits-cpu>`_ - Sets CPU and Memory resources allocated to members of the ReplSet.
