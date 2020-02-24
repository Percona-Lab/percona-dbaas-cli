==================================
Percona DBaaS Command Line Tool
==================================

The Percona DBaaS Command Line Tool (percona-dbaas-cli) is a powerfull
instrument providing a unify way for the user to have a Database as a Service
(DBaaS) like experience, built on top of the Kubernetes-based cloud
infrastructure. It is a command-line interface makning a standardized way to
interact with Operators in Kubernetes. It works with the Percona Operators,
and is intended to be used with other databse-related Operators which will be
supported in the future.

Thus, percona-dbaas-cli acts as a middleware between the Kubernetes Operators
and other tools to enable a DBaaS-like experience.

.. note:: This is an experimental release of the Percona DBaaS Command Line Tool,
          not intended yet for the production environments.


Overview
========

.. toctree::
   :maxdepth: 1

   system-requirements


Setting up
============

.. toctree::
   :maxdepth: 1

   installation
   kubernetes
   basic-usage

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
