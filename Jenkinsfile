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

  stages {
    stage('Prepare workspace') {
      steps {
        sh "/bin/echo -n ${env.GIT_BRANCH} > git-branch"
        sh "/bin/echo -n ${env.GIT_COMMIT_SHORT} > git-sha"
        sh "/bin/echo -n \$(git show -s --format=%ct) > git-buildstamp"
      }
    }

    stage('Build') {
      steps {
        script {
          image = docker.build("${env.IMAGE}:${env.GIT_BRANCH_TAG}-${env.GIT_COMMIT_SHORT}")
        }
      }
    }

    stage('Push') {
      when {
        expression { env.DOCKER_REPO }
      }
      steps {
        script {
          docker.withRegistry("https://${env.DOCKER_REPO}", env.DOCKER_REPO_CREDENTIALS) {
            image.push("${env.GIT_BRANCH_TAG}")
          }
          sh "docker rmi ${env.DOCKER_REPO}/${env.IMAGE}:${env.GIT_BRANCH_TAG}"
        }
      }
    }

    stage('Remove image') {
      steps {
        sh "docker rmi ${env.IMAGE}:${env.GIT_BRANCH_TAG}-${env.GIT_COMMIT_SHORT}"
      }
    }
  }
}
