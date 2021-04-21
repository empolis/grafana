pipeline {
  agent {
    label 'docker'
  }

  environment {
    IMAGE = 'empolis/grafana'
  }

  options {
    buildDiscarder(
      logRotator(numToKeepStr: '4', artifactNumToKeepStr: '4')
    )
  }

  stages {
    stage('Build') {
      steps {
        sh "docker build -t '${env.IMAGE}:${env.GIT_COMMIT}' --build-arg GIT_TAG=${env.GIT_COMMIT} ."
      }
    }

    stage('Push') {
      when {
        branch 'v7.3.x'
      }
      steps {
        script {
          docker.withRegistry("https://registry.industrial-analytics.cloud", "empolis-registry") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push("gamma")
          }
          docker.withRegistry("https://174193726300.dkr.ecr.eu-central-1.amazonaws.com", "ecr:eu-central-1:jenkins-eia-kba-stage-2") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push("gamma")
          }
          docker.withRegistry("https://187689283976.dkr.ecr.eu-central-1.amazonaws.com", "ecr:eu-central-1:jenkins-eia-kba-prod") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push("gamma")
          }
        }
      }
    }
  }
}
