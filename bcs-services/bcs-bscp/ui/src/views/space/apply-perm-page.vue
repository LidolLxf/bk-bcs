<template>
  <div class="apply-perm-page">
    <bk-exception type="403" scene="part">
      <div class="apply-description">
        <h4 class="title">{{ permTitle }}</h4>
        <template v-if="noSpaceViewPerm">
          <p class="tips">{{ t('你没有相应业务的访问权限，请前往申请相关业务权限') }}</p>
        </template>
        <bk-button text theme="primary" @click="handleApply">{{ t('申请权限') }}</bk-button>
      </div>
    </bk-exception>
  </div>
</template>
<script setup lang="ts">
  import { computed } from 'vue';
  import { storeToRefs } from 'pinia';
  import { useI18n } from 'vue-i18n';
  import useGlobalStore from '../../store/global';

  const { applyPermUrl, applyPermResource } = storeToRefs(useGlobalStore());
  const { t, locale } = useI18n();

  const noSpaceViewPerm = computed(() => {
    const index = applyPermResource.value.findIndex((item) => item.action === 'find_business_resource');
    return index > -1;
  });

  const permTitle = computed(() => {
    if (applyPermResource.value.length > 0) {
      const resource = applyPermResource.value[applyPermResource.value.length - 1];
      // 权限名称未做国际化，用action字段代替
      const permissionName = locale.value === 'en' ? ` ${resource.action} ` : resource.action_name;
      return `${t('无')}${permissionName}${t('权限')}`;
    }
    return t('无访问权限');
  });

  const handleApply = () => {
    window.open(applyPermUrl.value, '__blank');
  };
</script>
<style lang="scss" scoped>
  .apply-perm-page {
    padding-top: 100px;
    height: 100%;
    font-size: 12px;
  }
  .title {
    margin: 0 0 8px;
    font-size: 14px;
    color: #63656e;
    line-height: 22px;
  }
  .tips {
    margin: 0 0 8px;
    font-size: 12px;
    color: #979ba5;
  }
</style>
