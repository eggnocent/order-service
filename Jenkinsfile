pipeline {
  agent any

  environment {
    IMAGE_NAME = 'order-service'
    DOCKER_CREDENTIALS = credentials('docker-credential')
    GITHUB_CREDENTIALS = credentials('github-credential')
    SSH_KEY = credentials('ssh-key')
    HOST = credentials('host')
    USERNAME = credentials('username')
    CONSUL_HTTP_URL = credentials('consul-http-url')
    CONSUL_HTTP_TOKEN = credentials('consul-http-token')
    CONSUL_WATCH_INTERVAL_SECONDS = 60
  }

  stages {
    stage('Check Commit Message') {
      steps {
        script {
          def commitMessage = sh(
            script: "git log -1 --pretty=%B",
            returnStdout: true
          ).trim()

          echo "Commit Message: ${commitMessage}"
          if (commitMessage.contains("[skip ci]")) {
            echo "Skipping pipeline due to [skip ci] tag in commit message."
            currentBuild.result = 'ABORTED'
            currentBuild.delete()
            return
          }

          echo "Pipeline will continue. No [skip ci] tag found in commit message."
        }
      }
    }

    stage('Set Target Branch') {
      steps {
        script {
          echo "GIT_BRANCH: ${env.GIT_BRANCH}"
          if (env.GIT_BRANCH == 'origin/master') {
            env.TARGET_BRANCH = 'master'
          } else if (env.GIT_BRANCH == 'origin/development') {
            env.TARGET_BRANCH = 'development'
          }

          echo "TARGET_BRANCH: ${env.TARGET_BRANCH}"
        }
      }
    }

    stage('Checkout Code') {
      steps {
        script {
          def repoUrl = 'https://github.com/eggnocent/order-service.git'

          checkout([$class: 'GitSCM',
            branches: [
              [name: "*/${env.TARGET_BRANCH}"]
            ],
            userRemoteConfigs: [
              [url: repoUrl, credentialsId: 'github-credential']
            ]
          ])

          sh 'ls -lah'
        }
      }
    }

    stage('Login to Docker Hub') {
      steps {
        script {
          withCredentials([usernamePassword(credentialsId: 'docker-credential', passwordVariable: 'DOCKER_PASSWORD', usernameVariable: 'DOCKER_USERNAME')]) {
            sh """
            echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin
            """
          }
        }
      }
    }

    stage('Build and Push Docker Image') {
      steps {
        script {
          def runNumber = currentBuild.number
          sh "docker build -t ${DOCKER_CREDENTIALS_USR}/${IMAGE_NAME}:${runNumber} ."
          sh "docker push ${DOCKER_CREDENTIALS_USR}/${IMAGE_NAME}:${runNumber}"
        }
      }
    }

    stage('Update docker-compose.yaml') {
      steps {
        script {
          def runNumber = currentBuild.number
          sh "sed -i 's|image: ${DOCKER_CREDENTIALS_USR}/${IMAGE_NAME}:[0-9]\\+|image: ${DOCKER_CREDENTIALS_USR}/${IMAGE_NAME}:${runNumber}|' docker-compose.yaml"
        }
      }
    }

    stage('Commit and Push Changes') {
      steps {
        script {
          sh """
          git config --global user.name 'Jenkins CI'
          git config --global user.email 'jenkins@company.com'
          git remote set-url origin https://${GITHUB_CREDENTIALS_USR}:${GITHUB_CREDENTIALS_PSW}@github.com/eggnocent/order-service.git
          git add docker-compose.yaml
          git commit -m 'Update image version to ${TARGET_BRANCH}-${currentBuild.number} [skip ci]' || echo 'No changes to commit'
          git pull origin ${TARGET_BRANCH} --rebase
          git push origin HEAD:${TARGET_BRANCH}
          """
        }
      }
    }

    stage('Deploy to Remote Server') {
    steps {
        script {
            withCredentials([
                string(credentialsId: 'consul-http-token', variable: 'CONSUL_HTTP_TOKEN'),
                string(credentialsId: 'consul-http-url', variable: 'CONSUL_HTTP_URL'),
                sshUserPrivateKey(credentialsId: 'ssh-key', keyFileVariable: 'SSH_KEY_FILE', usernameVariable: 'SSH_USER'),
                string(credentialsId: 'host', variable: 'HOST')
            ]) {
                def targetBranch = env.TARGET_BRANCH ?: 'master'
                def consulWatchInterval = env.CONSUL_WATCH_INTERVAL_SECONDS ?: '60'
                
                sh '''
                    set -e
                    
                    echo "=== Testing SSH Connection ==="
                    ssh -o StrictHostKeyChecking=no -i "$SSH_KEY_FILE" ${SSH_USER}@${HOST} "echo 'SSH Connection: OK'"
                    
                    echo "=== Deploying to Remote Server ==="
                    ssh -o StrictHostKeyChecking=no -i "$SSH_KEY_FILE" ${SSH_USER}@${HOST} bash <<'ENDSSH'
                        set -e
                        
                        TARGET_DIR="/home/eggnocent/mini-soccer-project/order-service"
                        
                        if [ -d "$TARGET_DIR/.git" ]; then
                            echo "Directory exists. Pulling latest changes..."
                            cd "$TARGET_DIR"
                            git pull origin "''' + targetBranch + '''"
                        else
                            echo "Directory does not exist. Cloning repository..."
                            mkdir -p "$(dirname "$TARGET_DIR")"
                            git clone -b "''' + targetBranch + '''" https://github.com/eggnocent/order-service.git "$TARGET_DIR"
                            cd "$TARGET_DIR"
                        fi
                        
                        echo "Setting up .env file..."
                        [ -f .env.example ] && cp .env.example .env || touch .env
                        
                        sed -i "s|^TIMEZONE=.*|TIMEZONE=Asia/Jakarta|" .env
                        sed -i "s|^CONSUL_HTTP_URL=.*|CONSUL_HTTP_URL=${CONSUL_HTTP_URL}|" .env
                        sed -i "s|^CONSUL_HTTP_PATH=.*|CONSUL_HTTP_PATH=backend\\/order-service|" .env
                        sed -i "s|^CONSUL_HTTP_TOKEN=.*|CONSUL_HTTP_TOKEN=${CONSUL_HTTP_TOKEN}|" .env
                        sed -i "s|^CONSUL_WATCH_INTERVAL_SECONDS=.*|CONSUL_WATCH_INTERVAL_SECONDS=''' + consulWatchInterval + '''|" .env
                        
                        echo "Starting Docker Compose..."
                        sudo docker compose up -d --build --force-recreate
                        
                        echo "✅ Deployment completed successfully!"
ENDSSH
                    
                    echo "✅ Remote deployment finished."
                '''
            }
        }
    }
}
  }
}
