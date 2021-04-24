@Library('ese-ia-deployment-config') _

pipeline {
  agent {
    label 'docker'
  }

  environment {
    IMAGE = 'empolis/grafana'
    DOCKER_REPO = dockerRepo()
    DOCKER_REPO_CREDENTIALS = dockerRepoCredentials()
    GIT_COMMIT_SHORT = sh(
      script: "printf \$(git rev-parse --short ${env.GIT_COMMIT})",
      returnStdout: true
    )
    GIT_BRANCH_TAG = env.GIT_BRANCH.replace('/', '_')
  }

  options {
    buildDiscarder(
      logRotator(numToKeepStr: '4', artifactNumToKeepStr: '4')
    )
  }

  stages {
    stage('Build') {
      steps {
        sh "/bin/echo -n ${env.GIT_BRANCH} > git-branch"
        sh "/bin/echo -n ${env.GIT_COMMIT_SHORT} > git-sha"
        sh "/bin/echo -n `git show -s --format=%ct` > git-buildstamp"
        image = docker.build("${env.IMAGE}:${env.GIT_COMMIT_SHORT}")
      }
    }

    stage('Push') {
      when {
        expression { env.DOCKER_REPO }
      }
      steps {
        script {
          docker.withRegistry("https://${env.DOCKER_REPO}", env.DOCKER_REPO_CREDENTIALS) {
            image.push("${env.GIT_COMMIT_SHORT}")
            image.push("${env.GIT_BRANCH_TAG}")
          }
        }
      }
    }
  }
}
