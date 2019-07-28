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

    stage('Tag') {
      when {
        branch 'v6.1.x'
      }
      steps {
        sh "docker tag '${env.IMAGE}:${env.GIT_COMMIT}' '${env.IMAGE}:latest'"
      }
    }

    stage('Push') {
      when {
        branch 'v6.1.x'
      }
      steps {
        script {
          docker.withRegistry("https://402631066376.dkr.ecr.eu-central-1.amazonaws.com", "ecr:eu-central-1:JenkinsAWS") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:latest").push()
          }
          docker.withRegistry("https://402631066376.dkr.ecr.eu-west-1.amazonaws.com", "ecr:eu-west-1:JenkinsAWS") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:latest").push()
          }
          docker.withRegistry("https://187689283976.dkr.ecr.eu-central-1.amazonaws.com", "ecr:eu-central-1:jenkins-eia-kba-prod") {
            docker.image("${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.IMAGE}:latest").push()
          }
        }
      }
    }
  }
}
