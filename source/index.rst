==================================
Percona DBaaS Command Line Tool
==================================

The Percona DBaaS Command Line Tool (percona-dbaas-cli) is a powerful
instrument providing a unified way for the user to have a Database as a Service
(DBaaS) like experience, built on top of the Kubernetes-based cloud
infrastructure. It is a command-line interface to interact with Operators in
Kubernetes. 

Thus, percona-dbaas-cli acts as a middleware making a standardized
way to automate the deployment, management, and scaling of containerized
databases. It currently works with the Percona Operators, and is intended to be
used with other database-related Operators which will be supported in the future.

.. note:: This is an experimental release of the Percona DBaaS Command Line Tool,
          not intended yet for production environments.


Overview
========

.. toctree::
   :maxdepth: 1

   capabilities
   system-requirements
   basic-usage


Setting up
============

.. toctree::
   :maxdepth: 1

   installation
   kubernetes

MySQL mode
=============

.. toctree::
   :maxdepth: 1

   mysql-usage
   mysql-limitations
   mysql-options

MongoDB mode
=============

.. toctree::
   :maxdepth: 1

   mongodb-usage
   mongodb-limitations
   mongodb-options

Reference
=========

.. toctree::
  :maxdepth: 1

  Release Notes <release-notes/index.rst>
