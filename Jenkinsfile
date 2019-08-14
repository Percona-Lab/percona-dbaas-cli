void CreateCluster(String CLUSTER_PREFIX) {
    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
        sh """
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
            gcloud config set project $GCP_PROJECT
            gcloud container clusters create --zone us-central1-a $CLUSTER_NAME-${CLUSTER_PREFIX} --cluster-version 1.12 --machine-type n1-standard-4 --preemptible --num-nodes=3 --network=jenkins-vpc --subnetwork=jenkins-${CLUSTER_PREFIX}
            kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user jenkins@"$GCP_PROJECT".iam.gserviceaccount.com
        """
   }
}
void pushArtifactFile(String FILE_NAME) {
    withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
        sh """
            S3_PATH=s3://percona-jenkins-artifactory/\$JOB_NAME/\$(git rev-parse --short HEAD)
            aws s3 ls \$S3_PATH/${FILE_NAME} || :
            aws s3 cp --quiet ${FILE_NAME} \$S3_PATH/${FILE_NAME} || :
        """
    }
}

void popArtifactFile(String FILE_NAME) {
    withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
        sh """
            S3_PATH=s3://percona-jenkins-artifactory/\$JOB_NAME/\$(git rev-parse --short HEAD)
            aws s3 cp --quiet \$S3_PATH/${FILE_NAME} ${FILE_NAME} || :
        """
    }
}
void runTest(String TEST_NAME, String CLUSTER_PREFIX) {
    popArtifactFile("$GIT_BRANCH-$GIT_SHORT_COMMIT-$TEST_NAME")
    sh """
        if [ -f "$GIT_BRANCH-$GIT_SHORT_COMMIT-$TEST_NAME" ]; then
            echo Skip $TEST_NAME test
        else
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            ./e2e-tests/$TEST_NAME/run
            touch $GIT_BRANCH-$GIT_SHORT_COMMIT-$TEST_NAME
        fi
    """
    pushArtifactFile("$GIT_BRANCH-$GIT_SHORT_COMMIT-$TEST_NAME")

    sh """
        rm -rf $GIT_BRANCH-$GIT_SHORT_COMMIT-$TEST_NAME
    """
}
void installRpms() {
    sh """
        sudo yum install -y https://repo.percona.com/yum/percona-release-latest.noarch.rpm || true
        sudo percona-release enable-only tools
        sudo yum install -y percona-xtrabackup-80 jq | true
    """
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
        CLUSTER_NAME = sh(script: "echo jenkins-dbaas-${GIT_SHORT_COMMIT} | tr '[:upper:]' '[:lower:]'", , returnStdout: true).trim()
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
                installRpms()
                sh '''
                    if [ ! -d $HOME/google-cloud-sdk/bin ]; then
                        rm -rf $HOME/google-cloud-sdk
                        curl https://sdk.cloud.google.com | bash
                    fi

                    source $HOME/google-cloud-sdk/path.bash.inc
                    gcloud components update kubectl
                    gcloud version

                    curl -s https://storage.googleapis.com/kubernetes-helm/helm-v2.14.0-linux-amd64.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 -zvxpf -
                    curl -s -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 --wildcards -zxvpf - '*/oc'
                '''
            }
        }
        stage('Build docker image') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                 withCredentials([usernamePassword(credentialsId: 'hub.docker.com', passwordVariable: 'PASS', usernameVariable: 'USER')]) {
                    sh '''
                        DOCKER_TAG=perconalab/percona-xtradb-cluster-operator:master-broker-$VERSION
                        docker_tag_file='./results/docker/TAG'
                        mkdir -p $(dirname ${docker_tag_file})
                        echo ${DOCKER_TAG} > "${docker_tag_file}"

                        sg docker -c "
                            docker login -u '${USER}' -p '${PASS}'
                            export IMAGE=\$DOCKER_TAG
                            ./e2e-tests/build
                            docker logout
                        "
                    '''
                }
                stash includes: 'results/docker/TAG', name: 'IMAGE'
                archiveArtifacts 'results/docker/TAG'
            }
        }
        stage('Run Tests') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            parallel {
                stage('PSMDB Tests') {
                    steps {
                        CreateCluster('psmdb')
                        runTest('psmdb', 'psmdb')
                   }
                }
                stage('PXC Tests') {
                    steps {
                        CreateCluster('pxc')
                        runTest('pxc', 'pxc')
                    }
                }
            }
        }
    }
    post {
        always {
            script {
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        unstash 'IMAGE'
                        def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                        withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                            sh """
                                curl -v -X POST \
                                    -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                    -d "{\\"body\\":\\"PXC operator docker - ${IMAGE}\\"}" \
                                    "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                            """
                        }
                    }
                }
            }
            script {
                if (env.CHANGE_URL) {
                    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
                    sh '''
                        source $HOME/google-cloud-sdk/path.bash.inc
                        gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
                        gcloud config set project $GCP_PROJECT
                        gcloud container clusters delete --zone us-central1-a $CLUSTER_NAME-pxc $CLUSTER_NAME-psmdb
                        sudo docker rmi -f \$(sudo docker images -q) || true

                        sudo rm -rf $HOME/google-cloud-sdk
                    '''
                    }
                }
            }
            sh '''
                sudo rm -rf ./*
            '''
            deleteDir()
        }
    }
}
