List of Options
----------------------------------

Any valid option from the Percona Server for MongoDB Operator `Custom Resource <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html>`_ can be set using the ``--options`` flag.

Here is the list of the valid commonly used options:

* `replsets.[rs0].size <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-size>`_ - The number of members in the ReplSet. By default it's internally named rs0.  Can be any valid number, although we recommend odd numbers less than 15.
* `replsets.[rs0].poddisruptionbudget.maxunavailable <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-poddisruptionbudget-maxunavailable>`_ - The amount of members in the ReplSet which can become unavailable within Kubernetes constraints.
* `replsets.[rs0].affinity <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-affinity-antiaffinitytopologykey>`_ - Lets you set affinity rules.  For instance, you can use this for Multi-AZ deployments.
* `replsets.[rs0].arbiter.enabled <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-arbiter-enabled>`_ - Lets you turn on the MongoDB arbiter for 2-member ReplSets
* `replsets.[rs0].resources <https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html#replsets-resources-limits-cpu>`_ - Sets CPU and Memory resources allocated to members of the ReplSet.
