def skipBranchBulds = true
if ( env.CHANGE_URL ) {
    skipBranchBulds = false
}

pipeline {
    environment {
        CLOUDSDK_CORE_DISABLE_PROMPTS = 1
        GIT_SHORT_COMMIT = sh(script: 'git describe --always --dirty', , returnStdout: true).trim()
        VERSION = "${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}"
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
                    curl -s -L https://github.com/src-d/go-license-detector/releases/latest/download/license-detector.linux_amd64.gz \
                        | gunzip > license-detector
                    sudo mv license-detector /usr/local/bin/license-detector
                    sudo chmod +x /usr/local/bin/license-detector 
                '''
            }
        }
        stage('GoLicenseDetector test') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
               sh """
                   license-detector ${WORKSPACE} | awk '{print \$2}' | awk 'NF > 0' > license-detector-new || true
                   diff -u build/tests/license/compare/license-detector license-detector-new
               """
            }
        }
        stage('GoLicense test') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                sh '''
                    sg docker -c "
                         build/bin/build-source
                         build/bin/build-binary
                    "
                '''

                sh """
                    CLI_VERSION=\$(cat VERSION | grep percona-dbaas-cli | awk '{print \$2}')
                    golicense ${WORKSPACE}/tmp/binary/percona-dbaas-cli-\$CLI_VERSION/linux/percona-dbaas \
                        | grep -v 'license not found'  \
                        | awk '{print \$2}' | sort | uniq > golicense-new || true
                    diff -u build/tests/license/compare/golicense golicense-new
                """
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
