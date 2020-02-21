
Example Usage
**********************************

.. code:: bash

   ./percona-dbaas mysql create-db example

Example Output:

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

Lists all database instances or clusters currently present or provides details
about the database instance or cluster with the given name.

Example Usage: Listing Databases
*************************************

.. code:: bash

   ./percona-dbaas mysql describe-db

Example Output:

.. code:: text

   NAME                STATUS
   example             ready
   example2            ready

Example Usage: Getting Details on a Database
********************************************

.. code:: bash

   ./percona-dbaas mysql describe-db example

Example Output:

.. code:: bash

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

Changes any of the optional values associated to an existing database instance
or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql modify-db example --options="pxc.size=5"

Example Output:

.. code:: bash

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

``delete-db``

Deletes a database instance or cluster with the given name.

Example Usage
*************

.. code:: bash

   ./percona-dbaas mysql delete-db example

Example Output:

.. code:: bash

   ARE YOU SURE YOU WANT TO DELETE THE DATABASE 'example'? Yes/No
   ALL YOUR DATA WILL BE LOST. USE '--preserve-data' FLAG TO SAVE IT.
   yes
   Deleting........................[done]

