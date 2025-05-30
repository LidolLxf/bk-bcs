<template>
  <BcsContent>
    <template #header>
      <HeaderNav :list="navList"></HeaderNav>
    </template>
    <div class="h-[calc(100%-60px)]" v-bkloading="{ isLoading }">
      <bcs-tab class="node-pool-tab h-full">
        <bcs-tab-panel :label="$t('cluster.ca.nodePool.title.nodeConfig')" name="config">
          <NodeConfig
            :default-values="detailData"
            :schema="schema"
            :cluster="curCluster"
            is-edit
            ref="nodePoolConfigRef"
            v-if="!isLoading">
          </NodeConfig>
        </bcs-tab-panel>
        <bcs-tab-panel :label="$t('cluster.ca.nodePool.title.initConfig')" name="basic">
          <NodePoolInfo
            :default-values="detailData"
            :schema="schema"
            :cluster="curCluster"
            is-edit
            :show-footer="false"
            v-if="!isLoading"
            ref="nodePoolInfoRef">
          </NodePoolInfo>
        </bcs-tab-panel>
      </bcs-tab>
    </div>
    <div class="bcs-fixed-footer">
      <bcs-button
        theme="primary"
        class="min-w-[88px]"
        :loading="saveLoading"
        @click="handleEditNodePool">{{$t('generic.button.save')}}</bcs-button>
      <bcs-button @click="handleCancel">{{$t('generic.button.cancel')}}</bcs-button>
    </div>
  </BcsContent>
</template>
<script lang="ts">
import { computed, defineComponent, onMounted, ref, toRefs } from 'vue';

import NodeConfig from './node-config.vue';
import NodePoolInfo from './node-pool-info.vue';

import { mergeDeep } from '@/common/util';
import BcsContent from '@/components/layout/Content.vue';
import { useAppData } from '@/composables/use-app';
import { useFocusOnErrorField } from '@/composables/use-focus-on-error-field';
import $i18n from '@/i18n/i18n-setup';
import $router from '@/router/index';
import $store from '@/store/index';
import { useClusterList } from '@/views/cluster-manage/cluster/use-cluster';
import HeaderNav from '@/views/cluster-manage/components/header-nav.vue';

export default defineComponent({
  components: {
    BcsContent,
    HeaderNav,
    NodePoolInfo,
    NodeConfig,
  },
  props: {
    clusterId: {
      type: String,
      default: '',
      required: true,
    },
    nodeGroupID: {
      type: String,
      default: '',
    },
  },
  setup(props) {
    const { clusterId, nodeGroupID } = toRefs(props);

    const detailData = ref<any>(null);
    const nodePoolInfoRef = ref<any>(null);
    const nodePoolConfigRef = ref<any>(null);
    const { clusterList } = useClusterList();
    const curCluster = computed(() => ($store.state as any).cluster.clusterList
      ?.find(item => item.clusterID === clusterId.value) || {});
    const navList = computed(() => [
      {
        title: clusterList.value.find(item => item.clusterID === clusterId.value)?.clusterName,
        link: {
          name: 'clusterMain',
        },
      },
      {
        title: 'Cluster Autoscaler',
        link: {
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            clusterId: clusterId.value,
          },
        },
      },
      {
        title: `${detailData.value?.nodeGroupID} (${detailData.value?.name}) `,
        link: {
          name: 'nodePoolDetail',
          params: {
            clusterId: props.clusterId,
            nodeGroupID: props.nodeGroupID,
          },
        },
      },
      {
        title: $i18n.t('cluster.ca.nodePool.action.edit'),
        link: null,
      },
    ]);

    const isLoading = ref(true);
    // 获取默认值
    const schema = ref({});
    const handleGetSchemaData = async () => {
      const data = await $store.dispatch('clustermanager/resourceSchema', {
        $cloudID: 'selfProvisionCloud',
        $name: 'nodegroup',
      });
      schema.value = data?.schema || {};
    };

    // 获取详情
    const handleGetNodeGroupDetail = async () => {
      const data = await $store.dispatch('clustermanager/nodeGroupDetail', {
        $nodeGroupID: nodeGroupID.value,
      });
      return data;
    };
    const { focusOnErrorField } = useFocusOnErrorField();

    // 保存
    const user = computed(() => $store.getters.user);
    const { getUserInfo } = useAppData();
    const saveLoading = ref(false);
    const handleEditNodePool = async () => {
      const nodePoolInfoValidate = await nodePoolInfoRef.value?.validate();
      const nodePoolConfigValidate = await nodePoolConfigRef.value?.validate();
      if (!nodePoolConfigValidate) {
        focusOnErrorField();
      }
      if (!nodePoolInfoValidate || !nodePoolConfigValidate) return;

      saveLoading.value = true;
      if (!user.value.username) {
        // 偶现获取用户信息失败问题
        console.warn('user not login, get user info', user.value);
        await getUserInfo();
      }
      const nodePoolData = nodePoolInfoRef.value?.getNodePoolData();
      const nodeConfigData = nodePoolConfigRef.value?.getNodePoolData();
      const data = {
        ...mergeDeep({
          nodeTemplate: {
            module: detailData.value.nodeTemplate?.module || {},
          },
        }, nodeConfigData, nodePoolData),
        $nodeGroupID: detailData.value.nodeGroupID,
        clusterID: curCluster.value.clusterID,
        region: curCluster.value.region,
        updater: user.value.username,
      };
      const result = await $store.dispatch('clustermanager/updateNodeGroup', data);
      saveLoading.value = false;
      if (result) {
        $router.push({
          name: 'clusterMain',
          query: {
            active: 'autoscaler',
            clusterId: clusterId.value,
          },
        });
      }
    };
    const handleCancel = () => {
      $router.back();
    };
    onMounted(async () => {
      isLoading.value = true;
      await handleGetSchemaData();
      detailData.value = await handleGetNodeGroupDetail();
      if (!detailData.value.nodeTemplate?.dataDisks?.length) {
        detailData.value.nodeTemplate.dataDisks = detailData.value.launchTemplate.dataDisks.map((item, index) => ({
          diskType: item.diskType,
          diskSize: item.diskSize,
          fileSystem: item.fileSystem || 'ext4',
          autoFormatAndMount: true,
          mountTarget: item.mountTarget || index > 0 ? `/data${index}` : '/data',
        }));
      }
      isLoading.value = false;
    });
    return {
      saveLoading,
      curCluster,
      isLoading,
      schema,
      detailData,
      navList,
      nodePoolInfoRef,
      nodePoolConfigRef,
      handleCancel,
      handleEditNodePool,
    };
  },
});
</script>
<style lang="postcss" scoped>

>>> .node-pool-tab {
  .bk-tab-section {
    padding: 0;
    height: calc(100% - 42px);
    overflow: auto;
    .bk-tab-content {
      height: 100%;
    }
  }
  .node-config {
    margin-bottom: 0;
    max-height: unset;
  }
}
</style>
