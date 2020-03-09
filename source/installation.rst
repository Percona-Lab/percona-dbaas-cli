Installation of the percona-dbaas-cli tool
==========================================

Source code for the Percona DBaaS CLI Tool is currently available on github as an
experimental repository inside of the Percona-Lab organization.

Percona provides dbaas-cli tool packages for automatic installation from software
repositories for the following GNU/Linux distributions:

* DEB packages for Debian based distributions such as Ubuntu,
* RPM packages for Red Hat based distributions such as CentOS.

Installing on Debian or Ubuntu
-----------------------------------

1. Configure Percona repositories using the `percona-release <https://www.percona.com/doc/percona-repo-config/percona-release.html>`_ tool. First you’ll need to download and install the official percona-release package from Percona::

     wget https://repo.percona.com/apt/percona-release_latest.generic_all.deb
     sudo dpkg -i percona-release_latest.generic_all.deb

#. Enable the testing component of the tools repository as follows::

         sudo percona-release enable tools testing

   See `percona-release official documentation <https://www.percona.com/doc/percona-repo-config/percona-release.html>`_ for details.

#. Install the ``percona-dbaas-cli`` package::

     sudo apt-get update
     sudo apt-get install percona-dbaas-cli

#. Once Percona DBaaS CLI Tool is installed you can run``percona-dbaas --help``
   to see the brief information about the tool commands.

Installing on  Red Hat and CentOS
-------------------------------------

1. Configure Percona repositories using the `percona-release <https://www.percona.com/doc/percona-repo-config/percona-release.html>`_ tool. First you’ll need to download and install the official percona-release package from Percona::

     sudo yum install https://repo.percona.com/yum/percona-release-latest.noarch.rpm

#. Enable the testing component of the tools repository as follows::

         sudo percona-release enable tools testing

   See `percona-release official documentation <https://www.percona.com/doc/percona-repo-config/percona-release.html>`_ for details.

#. Install the ``percona-dbaas-cli`` package::

      yum install percona-dbaas-cli

#. Once Percona DBaaS CLI Tool is installed you can run``percona-dbaas --help``
   to see the brief information about the tool commands.

