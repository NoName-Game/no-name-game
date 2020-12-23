node {
    def app
    def scmVars

   stage('Setup Workspace') {
       cleanWs()
       sh 'printenv'
   }

    stage('Clone repository') {
        scmVars = checkout scm
        sh 'ls -al'

        sh 'git clone  -b develop git@bitbucket.org:no-name-game/nn-grpc.git ./nn-grpc'
    }

    stage('Build image') {
        image_id = sh (script: 'docker images -q valkyrie00/nn-telegram:${BRANCH_NAME#*/}', returnStdout: true).trim()
        if (image_id.isEmpty()) {
            app = docker.build('valkyrie00/nn-telegram:${BRANCH_NAME#*/}', '-f deployment/docker/Dockerfile .')
            docker.withRegistry('', '94376016-b8fd-4049-b17f-df423b6c5611') {
                app.push('${BRANCH_NAME#*/}')
            }
        }
    }

    stage('Deploy') {
        sh 'sed -i s/BRANCH_NAME/${BRANCH_NAME#*/}/g deployment/k8s/*'
        withKubeConfig([credentialsId: 'ca75fb51-272e-435e-86fb-cbbd942d270f']) {
              sh 'kubectl apply -f deployment/k8s/deployment-client.yml'
              sh 'kubectl apply -f deployment/k8s/deployment-notifier.yml'
        }
    }
}