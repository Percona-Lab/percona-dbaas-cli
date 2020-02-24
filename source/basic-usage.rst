Basic Usage:
==================================

The basic form of the command looks as follows:

.. code:: bash

   percona-dbaas <engine> <subcommand> name [--optional-parameters]

Three mandatory parameters are *engine*, *subcommand* and *name*.

Engine specifies the family of databases the command will deal with. Currently,
two engines are supported, and here they are:

* ``mysql`` - allows to manage MySQL databases,
* ``mongodb`` - allows to manage MongoDB databases.

Subcommand is specific to the engine and defines the action which should be done
(creating new databse or modifying an existing one, etc.).

Fnially, the name is used as a subcommand parameter to specify the object under action.
