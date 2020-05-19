void CreateCluster(String CLUSTER_PREFIX) {
    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
        echo "Create cluster"
        sh """
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
            gcloud config set project $GCP_PROJECT
            gcloud container clusters create --zone us-central1-a $CLUSTER_NAME-${CLUSTER_PREFIX} --cluster-version 1.15 --machine-type n1-standard-4 --preemptible --num-nodes=3 --network=jenkins-vpc --subnetwork=jenkins-${CLUSTER_PREFIX} --no-enable-autoupgrade
            kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user jenkins@"$GCP_PROJECT".iam.gserviceaccount.com
        """
   }
}

void ShutdownCluster(String CLUSTER_PREFIX) {
     echo "Shutdown cluster"
    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
        sh """
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
            gcloud config set project $GCP_PROJECT
            gcloud container clusters delete --zone us-central1-a $CLUSTER_NAME-${CLUSTER_PREFIX}
        """
   }
}

void RunTests(String CLUSTER_PREFIX) {
    echo "Start cli tests"
    try {
        echo "Start tests try"
        sh """
            sudo chmod +x ./dbaas-cli/integtests/run.sh
            sg docker -c "
                        docker run \
                            --rm \
                            -v $WORKSPACE:/go/src/github.com/Percona-Lab/percona-dbaas-cli \
                            -w /go/src/github.com/Percona-Lab/percona-dbaas-cli \
                            -e GO111MODULE=on \
                            golang:1.13 ./dbaas-cli/integtests/run.sh
                    "
            sudo chmod +x ./integtests
            sudo chmod +x ./percona-dbaas
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            ./integtests ./percona-dbaas
        """
    }
    catch (exc) {
        currentBuild.result = 'FAIL'
    }
    echo "The test was finished!"
}

def skipBranchBulds = true
if ( env.CHANGE_URL ) {
    skipBranchBulds = false
}

pipeline {
    environment {
        CLOUDSDK_CORE_DISABLE_PROMPTS = 1
        GIT_SHORT_COMMIT = sh(script: 'git describe --always --dirty', , returnStdout: true).trim()
        VERSION = "${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}"
        CLUSTER_NAME = sh(script: "echo jenkins-pxc-${GIT_SHORT_COMMIT} | tr '[:upper:]' '[:lower:]'", , returnStdout: true).trim()
        AUTHOR_NAME  = sh(script: "echo ${CHANGE_AUTHOR_EMAIL} | awk -F'@' '{print \$1}'", , returnStdout: true).trim()
    }
    agent {
        label 'docker'
    }
    stages {
        stage('Prepare') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                script {
                    if ( AUTHOR_NAME == 'null' )  {
                        AUTHOR_NAME = sh(script: "git show -s --pretty=%ae | awk -F'@' '{print \$1}'", , returnStdout: true).trim()
                    }  
                }
                sh '''
                    curl -s -L https://github.com/mitchellh/golicense/releases/latest/download/golicense_0.2.0_linux_x86_64.tar.gz \
                        | sudo tar -C /usr/local/bin --wildcards -zxvpf -
                #    curl -s -L https://github.com/src-d/go-license-detector/releases/latest/download/license-detector.linux_amd64.gz \
                #        | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    curl -s -L https://github.com/src-d/go-license-detector/releases/download/v3.0.2/license-detector.linux_amd64.gz \
                        | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    sudo chmod +x /usr/local/bin/license-detector 
                '''
                 sh '''
                    if [ ! -d $HOME/google-cloud-sdk/bin ]; then
                        rm -rf $HOME/google-cloud-sdk
                        curl https://sdk.cloud.google.com | bash
                    fi
                    source $HOME/google-cloud-sdk/path.bash.inc
                    gcloud components install alpha
                    gcloud components install kubectl
                
                    curl -s https://storage.googleapis.com/kubernetes-helm/helm-v2.16.1-linux-amd64.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 -zvxpf -
                    curl -s -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 --wildcards -zxvpf - '*/oc'
                    
                    curl -s -L https://github.com/mitchellh/golicense/releases/latest/download/golicense_0.2.0_linux_x86_64.tar.gz \
                        | sudo tar -C /usr/local/bin --wildcards -zxvpf -
                  #  curl -s -L https://github.com/src-d/go-license-detector/releases/latest/download/license-detector.linux_amd64.gz \
                  #      | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    curl -s -L https://github.com/src-d/go-license-detector/releases/download/v3.0.2/license-detector.linux_amd64.gz \
                        | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    sudo chmod +x /usr/local/bin/license-detector 
                '''
                withCredentials([file(credentialsId: 'cloud-secret-file', variable: 'CLOUD_SECRET_FILE')]) {
                    sh '''
                        cp $CLOUD_SECRET_FILE ./cloud-secret.yml
                    '''
                }
            }
        }
        stage('Run tests') {
            steps {
                CreateCluster('basic')
                RunTests('basic')
                ShutdownCluster('basic')
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                            sh """
                                curl -v -X POST \
                                    -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                    -d "{\\"body\\":\\"License check is ok. \\"}" \
                                    "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                            """
                        }
                    }
                 }
                 else {
                     slackSend channel: '#cloud-dev-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}, ${BUILD_URL} owner: @${AUTHOR_NAME}"
                 }
            }
            deleteDir()
        }
    }
}
