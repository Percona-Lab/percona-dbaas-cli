Using with Google Kubernetes Engine
====================================

Setting Up Kubernetes
----------------------------------

Here is a simple example showing the basic requirements to set up a Google
Kubernetes Engine (GKE) cluster in the Google Cloud, ready to deploy databases
to.

.. code:: bash

   gcloud auth login
   gcloud config set project your-project-name
   gcloud container clusters create --zone us-central1-a your-cluster-name --cluster-version 1.15 --num-nodes=3
   kubectl create clusterrolebinding cluster-admin-binding-$USER --clusterrole=cluster-admin --user=$(gcloud config get-value core/account)

After you have a running GKE cluster (or are prepared to use an existing one)
you need to create a Cluster Role Binding for Kubernetes to know that your
Google Cloud username is a valid Kubernetes cluster-admin.

.. code:: bash

   kubectl create clusterrolebinding cluster-admin-binding1 --clusterrole=cluster-admin --user=your.gcloud.user@gmail.com

Kubernetes Namespaces
----------------------------------

Create and switch context to your preferred namespace before running any
commands. By default, we'll use the built-in ``default`` namespace.  Namespaces
are a nice way to have multiple applications and databases share a single
Kubernetes cluster in a clean fashion.
