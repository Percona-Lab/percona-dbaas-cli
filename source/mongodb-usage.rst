Supported Sub-Commands
========================

.. contents::
   :local:
   :depth: 1


``create-db``
-------------

Creates a new databases instance or cluster with the given name.

Example Usage
*************

The following code should create the "example" MongoDB database:

.. code:: bash

   ./percona-dbaas mongodb create-db example

The output of this command should look as follows:

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

Lists all database instances or clusters currently present or provides details
about the database instance or cluster with the given name.


Example Usage: Listing Databases
********************************

.. code:: bash

   ./percona-dbaas mongodb describe-db

Example Output:

.. code:: bash

   NAME                STATUS
   example             ready
   example2            ready

Example Usage: Getting Details on a Database
********************************************

.. code:: bash

   ./percona-dbaas mongodb describe-db example2

Example Output:

.. code:: bash

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

Changes any of the optional values associated to an existing database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mongodb modify-db example --options="pxc.size=5"

Example Output:

.. code:: text

   


``delete-db``
-------------

Deletes a database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mongodb delete-db example

Example Output:


