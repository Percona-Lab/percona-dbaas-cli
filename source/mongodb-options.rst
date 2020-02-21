
List of Options
----------------------------------

Any valid option from the CustomResource can be set using the ``--options``
flag, a list of valid commonly used options is below:

* ``replsets.[rs0].size`` - The number of members in the ReplSet.  By default
  it's internally named rs0.  Can be any valid number, although we recommend
  odd numbers less than 15.
* ``replsets.[rs0].poddisruptionbudget.maxunavailable`` - The amount of members
  in the ReplSet which can become unavailable within Kubernetes constraints.
* ``replsets.[rs0].affinity`` - Lets you set affinity rules.  For instance, you
  can use this for Multi-AZ deployments.
* ``replsets.[rs0].arbiter.enabled`` - Lets you turn on the MongoDB arbiter for
  2-member ReplSets
* ``replsets.[rs0].resources`` - Sets CPU and Memory resources allocated to
  members of the ReplSet.
