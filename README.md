# library
- sarama(kafka library)
- viper
- gin gonic
- gorilla/websocket

# 폴더 구조
폴더만 설명합니다.
```shell
Root Project
C:.
│  config.yml               
│  Dockerfile               
│  go.mod
│  go.sum
│  main.go                  
│  README.md                
│
├─.idea
│      .gitignore
│      modules.xml
│      vcs.xml
│      websocket-go.iml
│      workspace.xml
│
├─configuration                              - config를 사용하도록 해주는 소스
│      cfg.go
│
├─helm                                       - helm 생성 폴더
│  │  websocket-go-chart-0.1.0.tgz
│  │
│  └─websocket-go
│      │  .helmignore
│      │  Chart.yaml
│      │  values.yaml
│      │
│      ├─charts
│      └─templates
│          │  config.yaml
│          │  deployment.yaml
│          │  hpa.yaml
│          │  ingress.yaml
│          │  NOTES.txt
│          │  service.yaml
│          │  serviceaccount.yaml
│          │  _helpers.tpl
│          │
│          └─tests
│                  test-connection.yaml
│
├─k8s                                        - k8s 배포 폴더
│      config.yaml
│      deployment.yaml
│
├─kafka                                      - kafka Consumer 소스
│      consumer.go
│      consumerGroup.go
│
└─server                                     - websocket 소스
        client.go
        hub.go
        websocket.go


```
 
## 1. config.yml을 생성
config.yml을 사용합니다. dbconfig를 사용하지 않고 kubernetes의 ENV나 docker ENV를 사용가능합니다.
```yaml
kafka:
  Topic: test-topic
  Broker: 172.30.40.20:9092,172.30.40.21:9092,172.30.40.22:9092
  Consumergroup: test-groups2
  InitialOffset: newest # oldest or newest
  assignor: roundrobin # roundrobin,sticky,range
  Oldest: false # true or false
```

## 2. kubernetes 배포 방법
[k8s](k8s)/[deployment.yaml](k8s%2Fdeployment.yaml)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: websocket-go
  labels:
    app: websocket-go
spec:
  replicas: 1
  selector:
    matchLabels:
      app: websocket-go
  template:
    metadata:
      labels:
        app: websocket-go
    spec:
      containers:
        - name: websocket-go
          image: sw90lee/websocket-go:0.1
          ports:
            - containerPort: 8080
          env:
            - name: Topic
              valueFrom:
                configMapKeyRef:
                  key: Topic
                  name: websocket-go-cm
            - name: Broker
              valueFrom:
                configMapKeyRef:
                  key: Broker
                  name: websocket-go-cm
            - name: Consumergroup
              valueFrom:
                configMapKeyRef:
                  key: Consumergroup
                  name: websocket-go-cm
            - name: InitialOffset
              valueFrom:
                configMapKeyRef:
                  key: InitialOffset
                  name: websocket-go-cm
            - name: assignor
              valueFrom:
                configMapKeyRef:
                  key: assignor
                  name: websocket-go-cm
            - name: Oldest
              valueFrom:
                configMapKeyRef:
                  key: Oldest
                  name: websocket-go-cm

---
apiVersion: v1
kind: Service
metadata:
  name: websocket-go-service
spec:
  selector:
    app: websocket-go
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  type: ClusterIP

---
kind: Ingress
apiVersion: networking.k8s.io/v1
metadata:
  name: websocket-go
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: websocket.icnp.in-soft.co.kr
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: websocket-go-service
                port:
                  number: 8080
```

[k8s](k8s)/[config.yaml](k8s%2Fdeployment.yaml)
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: websocket-go-cm
data:
  Topic: "test-topic"
  Broker: "172.30.40.20:9092,172.30.40.21:9092,172.30.40.22:9092"
  Consumergroup: "test-groups2"
  InitialOffset: "newest" # oldest or newest
  assignor: "roundrobin" # roundrobin,sticky,range
  Oldest: "false" # true or false
  
```
## 3. helm chart 
### 3-1. helm chart 만들기
helm chart는 helm/websocket-go/Chart.yaml을 기준으로 생성됩니다.
```shell
$ pwd
C:\Users\Insoft\GolandProjects\websocket-go\helm
# 문법적 오류 검사
$ helm list <Chart.yaml 경로>
# 배포되는 yaml 소스 확인
$ helm template <Chart.yaml 경로>
# chart 를 배포하기위한 tgz 파일 생성 (이름 버전은 chart.yaml을 따라갑니다.)
$ helm package <Chart.yaml 경로>
```
### 3-2. helm chart 배포하기

[helm](helm)/[websocket-go-chart-0.1.0.tgz](helm%2Fwebhook-go%2Fwebhook-go-chart-0.1.0.tgz)

- 해당 yaml의 환경변수 Image를 변경해서 배포합니다. 필요시 Ingress도 Enable로 진행하여 배포진행합니다.
```yaml
# Default values for websocket-go.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

# KAFKA 환경변수 작성
kafka:
  Topic: "test-topic"
  Broker: "172.30.40.20:9092,172.30.40.21:9092,172.30.40.22:9092"
  Consumergroup: "test-groups2"
  InitialOffset: "newest" # oldest or newest
  assignor: "roundrobin" # roundrobin,sticky,range
  Oldest: "false" # true or false

# Webhook IMAGE 작성
image:
  repository: sw90lee/websocket-go
  pullPolicy: IfNotPresent
  # Overrides the image tag whose default is the chart appVersion.
  tag: "0.1"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  # Specifies whether a service account should be created
  create: true
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name: "websocket-go"

podAnnotations: {}

podSecurityContext: {}
# fsGroup: 2000

securityContext: {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
# runAsUser: 1000
healthCheck:
  enabled: false

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: ""
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /
  hosts:
    - host: websocket.icnp.in-soft.co.kr # websocket 접속정보 url 작성
      paths:
        - path: /
          pathType: Prefix
  tls: []
  #  - secretName: chart-example-tls
  #    hosts:
  #      - chart-example.local
probe:
  enabled: false
  livenessProbe: {}
  #    httpGet:
  #      path: /actuator/health
  #      port: 8080
  #    timeoutSeconds: 1
  #    successThreshold: 1
  #    failureThreshold: 30
  #    periodSeconds: 10
  readinessProbe: {}
  #    httpGet:
  #      path: /actuator/health
  #      port: 8080
  #    timeoutSeconds: 1
  #    successThreshold: 1
  #    failureThreshold: 30
  #    periodSeconds: 10
  resources: {}
    # We usually recommend not to specify default resources and to leave this as a conscious
    # choice for the user. This also increases chances charts run on environments with little
    # resources, such as Minikube. If you do want to specify resources, uncomment the following
    # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
    # limits:
    #   cpu: 100m
    #   memory: 128Mi
    # requests:
    #   cpu: 100m
  #   memory: 128Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80
  # targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}

```


- Helm install 진행 
```bash
# helm 
$ helm install <helm name> websocket-go-chart-0.1.0.tgz -f value.yaml
```
