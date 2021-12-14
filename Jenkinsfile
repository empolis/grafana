@Library('ese-ia-deployment-config') _

pipeline {
    agent {
        label 'docker'
    }

    environment {
        IMAGE = 'eseia/grafana'
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
                expression { dockerRepos() && !(env.BRANCH_NAME =~ /release\/.+/) }
            }
            steps {
                script {
                    dockerRepos().each { repo ->
                        docker.withRegistry("https://${repo.dockerRepo}", repo.credentialsId) {
                            image.push("${env.GIT_BRANCH_TAG}")
                        }
                        sh "docker rmi ${repo.dockerRepo}/${env.IMAGE}:${env.GIT_BRANCH_TAG}"
                    }
                }
            }
        }

        stage('Push release image with build timestamp') {
            when {
                expression { dockerRepos() && env.BRANCH_NAME =~ /release\/.+/ }
            }
            steps {
                script {
                    dockerRepos().each { repo ->
                        docker.withRegistry("https://${repo.dockerRepo}", repo.credentialsId) {
                            image.push("${env.GIT_BRANCH_TAG}-${env.BUILD_TIMESTAMP}")
                        }
                        sh "docker rmi ${repo.dockerRepo}/${env.IMAGE}:${env.GIT_BRANCH_TAG}-${env.BUILD_TIMESTAMP}"
                    }
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
