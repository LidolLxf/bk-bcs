<template src="./tmpl-list.html"></template>

<script>
import State from '../k8s-state';
import mixinBaseList from '../mixins/mixin-base-list';

import BcsContent from '@/components/layout/Content.vue';
import Row from '@/components/layout/Row.vue';

export default {
  name: 'K8SDeployments',
  components: { BcsContent, Row },
  mixins: [mixinBaseList],
  data() {
    return {
      CATEGORY: 'deployment',
      State,
    };
  },
  methods: {
    /**
             * 跳转到模板实例化页面
             *
             * @param {Object} tmplMuster 当前模板集对象
             * @param {Object} tpl 当前模板对象
             */
    goInstantiation(tmplMuster, tpl) {
      this.$router.push({
        name: 'deploymentsInstantiation',
        params: {
          projectId: this.projectId,
          projectCode: this.projectCode,
          templateId: tmplMuster.tmpl_muster_id,
          category: tpl.category,
          tmplAppId: tpl.tmpl_app_id,
          tmplAppName: tpl.tmpl_app_name,
          searchParamsList: this.searchParamsList,
        },
      });
    },

    /**
             * 跳转到实例详情页面
             *
             * @param {Object} instance 当前实例对象
             * @param {Object} namespace 当前 namespace 对象，只有命名空间试图才会有
             */
    async goInstanceDetail(instance, namespace) {
      const params = {
        projectId: this.projectId,
        projectCode: this.projectCode,
        instanceId: instance.id,
        templateId: instance.templateId,
        instanceName: instance.name,
        instanceNamespace: instance.namespace,
        instanceCategory: instance.category,
        searchParamsList: this.searchParamsList,
      };

      if (namespace) {
        params.namespaceId = namespace.id;
      }

      this.$router.push({
        name: String(instance.id) === '0' ? 'deploymentsInstanceDetail2' : 'deploymentsInstanceDetail',
        params,
        query: {
          cluster_id: instance.cluster_id,
          name: params.instanceName,
          namespace: params.instanceNamespace,
        },
      });
    },
  },
};
</script>

<style scoped>
    @import '../list.css';
</style>
