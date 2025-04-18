{{- define "bscpconfig.Data" }}
spec:
  title: {{ i18n "配置信息" .lang }}
  type: object
  properties:
    provider:
      title: {{ i18n "provider" .lang }}
      type: object
      required:
        - feedAddr
        - biz
        - token      
        - app      
      properties:
        feedAddr: 
          title: {{ i18n "feedAddr" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "feed 地址" .lang }}
              maxRows: 6
          ui:rules:
            - required
        biz: 
          title: {{ i18n "biz" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "业务" .lang }}
              maxRows: 6
          ui:rules:
            - required
        token: 
          title: {{ i18n "token" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "客户端管理的token" .lang }}
              maxRows: 6
          ui:rules:
            - required
        app: 
          title: {{ i18n "app" .lang }}
          type: string
          ui:component:
            name: bfInput
            props:
              placeholder: {{ i18n "bscp app的名称" .lang }}
              maxRows: 6
          ui:rules:
            - required
      ui:component:
        props:
          visible: true
      ui:rules:
        - required
    configSyncer:
      title: {{ i18n "configSyncer" .lang }}
      type: array
      items:
        type: object
        required:
          - configmapName
          - matchConfigs
          - secretName      
          - secretType         
        properties:
          resourceType:
            title: {{ i18n "资源类型" .lang }}
            type: string
            default: configmapName
            ui:group:
              props:
                showTitle: false
            ui:component:
              name: radio
              props:
                datasource:
                  - label: {{ i18n "configmapName" .lang }}
                    value: configmapName
                  - label: {{ i18n "secretName" .lang }}
                    value: secretName
            ui:reactions:
              - target: "{{`{{`}} $widgetNode?.getSibling('configmapName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'configmapName' {{`}}`}}"
                then:
                  state:
                    visible: true
                    value: ""
                else:
                  state:
                    visible: false
                    value: "configmapName"
              - target: "{{`{{`}} $widgetNode?.getSibling('matchConfigs')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'configmapName' {{`}}`}}"
                then:
                  state:
                    visible: true
                    value: ""
                else:
                  state:
                    visible: false
                    value: "matchConfigs"          
              - target: "{{`{{`}} $widgetNode?.getSibling('secretName')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'secretName' {{`}}`}}"
                then:
                  state:
                    visible: true
                    value: ""
                else:
                  state:
                    visible: false
                    value: "secretName"
              - target: "{{`{{`}} $widgetNode?.getSibling('secretType')?.id {{`}}`}}"
                if: "{{`{{`}} $self.value === 'secretName' {{`}}`}}"
                then:
                  state:
                    visible: true
                else:
                  state:
                    visible: false
          configmapName:
            title: {{ i18n "configmapName" .lang }}
            type: string 
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "configmapName" .lang }}
                maxRows: 6
            ui:rules:
              - required       
          matchConfigs:
            title: {{ i18n "matchConfigs" .lang }}
            type: string
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "支持linux wilecard语法" .lang }}
                maxRows: 6
            ui:rules:
              - required              
          secretName:
            title: {{ i18n "secretName" .lang }}
            type: string
            ui:component:
              name: bfInput
              props:
                placeholder: {{ i18n "secretName" .lang }}
                maxRows: 6
            ui:rules:
              - required                  
          secretType:
            title: {{ i18n "type" .lang }}
            type: string  
            default: Opaque
            ui:component:
              name: select
              props:
                clearable: false
                datasource:
                  - label: Opaque
                    value: Opaque
                  - label: kubernetes.io/service-account-token
                    value: kubernetes.io/service-account-token
                  - label: kubernetes.io/dockerconfigjson
                    value: kubernetes.io/dockerconfigjson     
                  - label: kubernetes.io/basic-auth
                    value: kubernetes.io/basic-auth    
                  - label: kubernetes.io/ssh-auth
                    value: kubernetes.io/ssh-auth   
                  - label: kubernetes.io/tls
                    value: kubernetes.io/tls   
                  - label: bootstrap.kubernetes.io/token
                    value: bootstrap.kubernetes.io/token
          resourceData:
            title: {{ i18n "resourceData" .lang }}
            type: object
            properties:                                                                                     
              configData:
                title: {{ i18n "data" .lang }}
                type: array
                items:
                  type: object
                  required:
                    - key
                    - refConfig                  
                  properties:
                    key:
                      title: {{ i18n "key" .lang }}
                      type: string
                      ui:component:
                        name: bfInput
                        props:
                          placeholder: {{ i18n "key" .lang }}
                          maxRows: 6
                      ui:rules:
                        - required       
                    refConfig:
                      title: {{ i18n "refConfig" .lang }}
                      type: string
                      ui:component:
                        name: bfInput
                        props:
                          placeholder: {{ i18n "refConfig" .lang }}
                          maxRows: 6
                      ui:rules:
                        - required                      
                ui:group:
                  props:
                    showTitle: false
            ui:group:
              name: collapse
              props:
                border: false
                showTitle: false                    
        ui:order:
          - resourceType
          - configmapName
          - matchConfigs
          - secretName
          - secretType        
          - resourceData
      ui:group:
        props:
          showTitle: false
  ui:group:
    name: collapse
    props:
      border: true
      showTitle: true
      type: card
      verifiable: true
      hideEmptyRow: true
      defaultActiveName:
        - provider          
  ui:order:
    - provider
    - configSyncer            
{{- end }}
