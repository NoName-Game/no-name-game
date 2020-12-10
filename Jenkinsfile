node {
    def app
    def scmVars

   stage('Setup && Clean Workspace') {
       cleanWs()
       sh 'printenv'
   }

    stage('Clone repository') {
        scmVars = checkout scm
        sh 'git submodule update --init'
        sh 'ls -al'
    }

    stage('Build image') {
        app = docker.build("valkyrie00/nn-telegram:" + scmVars.GIT_COMMIT, "-f deployment/docker/Dockerfile .")
    }

    stage('Push image') {
        docker.withRegistry('', '94376016-b8fd-4049-b17f-df423b6c5611') {
            app.push(scmVars.GIT_COMMIT)
        }
    }

//     stage('Deploy Stagin Approval') {
//         timeout(time: 1, unit: 'MINUTES') {
//             input(id: "Deploy Staging Gate", message: "Deploy on staging?", ok: 'Ok Deploy!')
//         }
//     }

    stage('Deploy Staging') {
        sh 'sed -i s/GIT_COMMIT/'+scmVars.GIT_COMMIT+'/g deployment/k8s/*'
        withKubeConfig([credentialsId: 'ca75fb51-272e-435e-86fb-cbbd942d270f']) {
              sh 'kubectl apply -f deployment/k8s/deployment-client.yml'
        }
    }
}