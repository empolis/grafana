pipeline {
  agent {
    label 'docker'
  }

  environment {
    REGISTRY = '402631066376.dkr.ecr.eu-central-1.amazonaws.com'
    IMAGE = 'empolis/grafana'
  }

  options {
    buildDiscarder(
      logRotator(numToKeepStr: '4', artifactNumToKeepStr: '4')
    )
  }

  stages {
    stage('Build') {
      when {
        branch 'v6.0.x'
      }
      steps {
        sh "docker build -t '${env.REGISTRY}/${env.IMAGE}:${env.GIT_COMMIT}' --build-arg GIT_TAG=${env.GIT_COMMIT} ."
        sh "docker tag '${env.REGISTRY}${env.IMAGE}:${env.GIT_COMMIT}' '${env.REGISTRY}/${env.IMAGE}:latest'"
      }
    }

    stage('Push to registry') {
      when {
        branch 'v6.0.x'
      }
      steps {
        script {
          docker.withRegistry("https://${env.REGISTRY}", "ecr:eu-central-1:JenkinsAWS") {
            docker.image("${env.REGISTRY}/${env.IMAGE}:${env.GIT_COMMIT}").push()
            docker.image("${env.REGISTRY}/${env.IMAGE}:latest").push()
          }
        }
      }
    }
  }
}
