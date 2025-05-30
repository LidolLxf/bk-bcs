# BscpConfig

> Bscpconfig is used to define the structural configuration of BSCP.

## Using BscpConfig

Common bscpconfig configuration examples are as follows

````yaml
apiVersion: bk.tencent.com/v1alpha1
kind: BscpConfig
metadata:
  name: "bscpconfig-simple"
  namespace: default
  labels: # Grayscale publishing use
    region: "ap-guangzhou2"
spec:
  provider:
    feedAddr: "feed.bscp.example.com:9510"
    biz: 100 # business
    token: "xxx" # Client secret key
    app: "bcs-gateway-certs" # Name of BSCP app

  configSyncer:
    - configmapName: "" # Generate a fixed name configmap
      matchConfigs: ["*"] # Configuration item matching rules, using Linux wilecard syntax


    - configmapName: "" # Generate a fixed name configmap
      matchConfigs: ["*credentials*", "xxx"] # Configuration item matching rules, using Linux wilecard syntax

    - secretName: bcs-gateway-certs # Generate fixed name secret
      type: kubernetes.io/tls # Secret specifies the type. The default is opaque
      data:
        - key: tls.key # secret data.key name
          refConfig: uat_tls_key # Secret data.value, refconfig is the exact configuration item name
        - key: tls.crt
          refConfig: uat_tls_ca
````
