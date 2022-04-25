@Library(value = 'msaas-shared-lib', changelog = false) _

pipeline {
 agent {
   kubernetes {
     label "build-flippy"
     yamlFile 'KubernetesPods.yaml'
   }
 }

 environment {
    GIT_TOKEN = credentials('GIT_TOKEN')
    commitSha = sh(returnStdout: true, script: 'git rev-parse --short=5 HEAD').trim()
 }

  stages {
    stage('build & push') {
      steps {
        script {
          container('podman') {
            withEnv(["HUB=docker.intuit.com/services/mesh/flippy/service", "TAG=${commitSha}", "VERSION=0.1"]) {
              withCredentials([[$class: 'UsernamePasswordMultiBinding',
                credentialsId: "flippy_docker_login",
                usernameVariable: 'DOCKER_REGISTRY_CREDS_USR',
                passwordVariable: 'DOCKER_REGISTRY_CREDS_PSW']]) {
                  echo 'Login to registry..'
                  sh "podman login docker.intuit.com --username ${DOCKER_REGISTRY_CREDS_USR} --password ${DOCKER_REGISTRY_CREDS_PSW} --storage-driver=overlay"
                  sh "podman build --storage-driver=overlay --isolation=chroot --ulimit=nofile=1048576:1048576 --cgroup-manager=cgroupfs --events-backend=file -t $HUB/flippy:latest -t $HUB/flippy:$TAG --build-arg GIT_TOKEN=$GIT_TOKEN ."
                  sh "podman push $HUB/flippy:$TAG --storage-driver=overlay --cgroup-manager=cgroupfs --events-backend=file"
                }
            }
          }
        }
      }
    }

    stage('master build build & push') {
      when {
        branch 'master'
      }
      steps {
        script {
          container('podman') {
            withEnv(["HUB=docker.intuit.com/services/mesh/flippy/service", "TAG=${commitSha}", "VERSION=0.1"]) {
              withCredentials([[$class: 'UsernamePasswordMultiBinding',
                credentialsId: "flippy_docker_login",
                usernameVariable: 'DOCKER_REGISTRY_CREDS_USR',
                passwordVariable: 'DOCKER_REGISTRY_CREDS_PSW']]) {
                  echo 'Login to registry..'
                  sh "podman login docker.intuit.com --username ${DOCKER_REGISTRY_CREDS_USR} --password ${DOCKER_REGISTRY_CREDS_PSW} --storage-driver=overlay"
                  sh "podman push $HUB/flippy:latest --storage-driver=overlay --cgroup-manager=cgroupfs --events-backend=file"
                }
            }
          }
        }
      }
    }



    stage('CPD Certification') {
      when {
        branch 'master'
      }
      steps {
        container('cpd2') {
          intuitCPD2Podman([asset_id: "7209154696006625069", code_repo: "https://github.intuit.com/services-mesh/flippy.git"], "-i docker.intuit.com/services/mesh/flippy/service/flippy:latest --buildfile Dockerfile", "flippy_docker_login")
        }
      }
    }
  }
}
