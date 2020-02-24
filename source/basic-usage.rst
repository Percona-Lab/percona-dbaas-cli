Basic Usage
==================================

The basic form of the command looks as follows:

.. code:: bash

   percona-dbaas <engine> <subcommand> name [--optional-parameters]

The mandatory parameters are *engine* and *subcommand*.

*Engine* specifies the family of database services the command will deal with, and
is bound to a specific Kubernetes Operator. Currently, two engines are
supported:

* ``mysql`` - allows to manage MySQL databases within the Percona XtraDB Cluster via the `Percona XtraDB Cluster Operator <https://www.percona.com/doc/kubernetes-operator-for-pxc/index.html>`_,
* ``mongodb`` - allows to manage MongoDB databases via the `Percona Server for MongoDB Operator <percona.com/doc/kubernetes-operator-for-psmongodb/index.html>`_.

*Subcommand* is specific to the engine and defines the action which should be done
(creating new databse or modifying an existing one, etc.).

Finally, the *name* parameter typically follows subcommand to specify the object under action.
