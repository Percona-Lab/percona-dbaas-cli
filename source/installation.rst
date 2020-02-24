Installation of the percona-dbaas-cli
=====================================

The Percona DBaaS CLI Tool is currently available on github as an experimental
repository inside of Percona-Lab. Following steps are required to install it
from source:

#. Clone the ``github.com/Percona-Lab/percona-dbaas-cli`` repo with the
   following command:

   .. code:: bash

      git clone https://github.com/Percona-Lab/percona-dbaas-cli

#. Enter to the ``percona-dbaas-cli`` directory and run the ``build-source``
   script which downloads necessary prerequisites and creates the source
   tarball:
   
   .. code:: bash

      build/bin/build-source

#. Build the percona-dbaas-cli binary:

   .. code:: bash

      build/bin/build-binary

   .. note:: You will need running docker and enough priviledges to access it.

#. Optionally you can run other scripts to create DEB or RPM packages, etc.
   These scripts are located in the same subfolder and have self-explanatory
   names.
