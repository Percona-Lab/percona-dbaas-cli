Supported subcommands
========================

The following subcommands can be used with the ``mysql`` engine:

.. contents::
   :local:
   :depth: 1


``create-db``
---------------

This subcommand creates a new database instance or cluster with the given name.


Example Usage
**********************************

The following code should create the "example" database within the
Percona XtraDB Cluster:

.. code:: bash

   ./percona-dbaas mysql create-db example

The output of the above command should look as follows:

.. code:: text

   Starting..................................[done]
   Database started successfully, connection details are below:
   Provider:          k8s
   Engine:            pxc
   Resource Name:     example
   Resource Endpoint: example-proxysql.example.pxc.svc.local
   Port:              3306
   User:              root
   Pass:              random-password
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example-proxysql 3306:3306 &
   mysql -h 127.0.0.1 -P 3306 -uroot -prandom-password


``describe-db``
---------------

This subcommand lists all database instances or clusters currently present or
provides details about the database instance or cluster with the given name.

Example Usage: Listing Databases
*************************************

.. code:: bash

   ./percona-dbaas mysql describe-db

The output of the above command should look as follows:

.. code:: text

   NAME                STATUS
   example             ready
   example2            ready

Example Usage: Getting Details on a Database
********************************************

.. code:: bash

   ./percona-dbaas mysql describe-db example

The output of the above command should look as follows:

.. code:: text

   Provider:          k8s
   Engine:            pxc
   Resource Name:     example
   Resource Endpoint: example-proxysql.example.pxc.svc.local
   Port:              3306
   User:              root
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example-proxysql 3306:3306 &
   mysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD

``modify-db``
---------------

This subcommand changes any of the optional values associated to an existing database instance
or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql modify-db example --options="pxc.size=5"

The output of the above command should look as follows:

.. code:: text

   Modifying..........................[done]
   Database modified successfully, connection details are below:
   Provider:          k8s
   Engine:            pxc
   Resource Name:     example
   Resource Endpoint: example-proxysql.example.pxc.svc.local
   Port:              3306
   User:              root
   Status:            ready

   To access database please run the following commands:
   kubectl port-forward svc/example-proxysql 3306:3306 &
   mysql -h 127.0.0.1 -P 3306 -uroot -pPASSWORD

``restart-db``
---------------

This subcommand restarts an already existing MySQL cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql restart-db example

The output of the above command should look as follows:

.. code:: text

   ARE YOU SURE YOU WANT TO RESTART THE DATABASE 'example'? Yes/No
   
   yes
   Restarting........................[done]

``stop-db``
---------------

This subcommand stops an already running MySQL cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql stop-db example

The output of the above command should look as follows:

.. code:: text

   ARE YOU SURE YOU WANT TO STOP THE DATABASE 'example'? Yes/No
   
   yes
   Stopping........................[done]

``start-db``
---------------

This subcommand starts an already existing MySQL cluster with the given name,
which was previously stopped with the ``stop-db`` command.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql start-db example

The output of the above command should look as follows:

.. code:: text

   ARE YOU SURE YOU WANT TO START THE DATABASE 'example'? Yes/No
   
   yes
   Starting........................[done]

``delete-db``
---------------

This subcommand deletes a database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql delete-db example

The output of the above command should look as follows:

.. code:: text

   ARE YOU SURE YOU WANT TO DELETE THE DATABASE 'example'? Yes/No
   ALL YOUR DATA WILL BE LOST. USE '--preserve-data' FLAG TO SAVE IT.
   yes
   Deleting........................[done]

