@Library('ese-ia-deployment-config') _

pipeline {
    agent {
        label 'docker'
    }

    environment {
        IMAGE = 'eseia/grafana'
        GIT_COMMIT_SHORT = sh(
                script: "printf \$(git rev-parse --short ${env.GIT_COMMIT})",
                returnStdout: true
        )
        GIT_BRANCH_TAG = env.GIT_BRANCH.replace('/', '_')
        BUILD_DATE = sh(
                script: "printf \"\$(date --rfc-3339=seconds --utc)\"",
                returnStdout: true
        )
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
                    image = docker.build("$IMAGE:$GIT_BRANCH_TAG-$GIT_COMMIT_SHORT", """\
                        --pull \
                        --label org.opencontainers.image.vendor="Empolis Information Management GmbH" \
                        --label org.opencontainers.image.title=$IMAGE \
                        --label org.opencontainers.image.created="$BUILD_DATE" \
                        --label org.opencontainers.image.revision=$GIT_BRANCH_TAG-$GIT_COMMIT_SHORT \
                        .""".stripIndent())
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
