Supported subcommands
========================

The following subcommands can be used with the ``mongodb`` engine:

.. contents::
   :local:
   :depth: 1

``create-db``
-------------

This subcommand creates a new database instance or cluster with the given name.

Example Usage
*************

The following code should create the "example" MongoDB database:

.. code:: bash

   ./percona-dbaas mongodb create-db example

The output of the above command should look as follows:

.. code:: text

   Starting.............................[done]
   Database started successfully, connection details are below:
   Provider:          k8s
   Engine:            psmdb
   Resource Name:     example
   Resource Endpoint: example.example.psmdb.svc.local
   Port:              27017
   User:              clusterAdmin
   Pass:              random-password
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example-example 27017:27017 &
   mongo mongodb://clusterAdmin:random-password@localhost:27017/admin?ssl=false

``describe-db``
---------------

This subcommand lists all database instances or clusters currently present or
provides details about the database instance or cluster with the given name.


Example Usage: Listing Databases
********************************

.. code:: bash

   ./percona-dbaas mongodb describe-db

The output of the above command should look as follows:

.. code:: text

   NAME                STATUS
   example             ready
   example2            ready

Example Usage: Getting Details on a Database
********************************************

.. code:: bash

   ./percona-dbaas mongodb describe-db example2

The output of the above command should look as follows:

.. code:: text

   Provider:          k8s
   Engine:            psmdb
   Resource Name:     example2
   Resource Endpoint: example2.example2.psmdb.svc.local
   Port:              27017
   User:              clusterAdmin
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example2-example2 27017:27017 &
   mongo mongodb://clusterAdmin:PASSWORD@localhost:27017/admin?ssl=false

``modify-db``
-------------

This subcommand changes any of the optional values associated to an existing
database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mongodb modify-db example --options="pxc.size=5"

The output of the above command should look as follows:

.. code:: text

   Modifying..........................[done]
   Database modified successfully, connection details are below:
   Provider:          k8s
   Engine:            psmdb
   Resource Name:     example
   Resource Endpoint: example.example.psmdb.svc.local
   Port:              27017
   User:              clusterAdmin
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example2-example2 27017:27017 &
   mongo mongodb://clusterAdmin:PASSWORD@localhost:27017/admin?ssl=false

``delete-db``
-------------

This subcommand deletes a database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mongodb delete-db example

The output of the above command should look as follows:

.. code:: text

   ARE YOU SURE YOU WANT TO DELETE THE DATABASE 'example'? Yes/No
   ALL YOUR DATA WILL BE LOST. USE '--preserve-data' FLAG TO SAVE IT.
   yes
   Deleting........................[done]

.. note:: You can use this subcommand with the additional ``--preserve-data``
   key to preventing it from deleting persistent volumes with the actual data.
   This volume may be later used with some Pod to access the data, e.g., to copy
   it with the ``kubectl cp`` command. Also, you can later create a new cluster
   with the same name, and volumes will be re-used so that the previous data
   will be present for MongoDB.
