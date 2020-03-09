Basic Usage
==================================

The basic form of the command looks as follows:

.. code:: bash

   percona-dbaas <engine> <subcommand> name [--optional-parameters]

The mandatory parameters are *engine* and *subcommand*.

*Engine* specifies the family of database services the command will deal with.
Currently, two engines are supported:

* ``mysql`` - allows to manage MySQL databases via the `Percona XtraDB Cluster Operator <https://www.percona.com/doc/kubernetes-operator-for-pxc/index.html>`_,
* ``mongodb`` - allows to manage MongoDB databases via the `Percona Server for MongoDB Operator <percona.com/doc/kubernetes-operator-for-psmongodb/index.html>`_.

*Subcommand* is specific to the engine and defines the action which should be done
(creating a new database or modifying an existing one, etc.).

Finally, the *name* parameter typically follows subcommand to specify the object under action.
