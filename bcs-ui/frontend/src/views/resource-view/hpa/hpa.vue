<template>
  <BaseLayout title="HPA" kind="HorizontalPodAutoscaler" type="hpa">
    <template
      #default="{
        curPageData, pageConf,
        handlePageChange, handlePageSizeChange,
        handleGetExtData, handleSortChange,
        handleShowDetail, handleUpdateResource,
        handleDeleteResource, handleShowViewConfig,
        clusterNameMap, goNamespace, isViewEditable,isClusterMode, sourceTypeMap
      }">
      <bk-table
        :data="curPageData"
        :pagination="pageConf"
        @page-change="handlePageChange"
        @page-limit-change="handlePageSizeChange"
        @sort-change="handleSortChange">
        <bk-table-column :label="$t('generic.label.name')" min-width="150" prop="metadata.name" sortable fixed="left">
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable"
              @click="handleShowDetail(row)">{{ row.metadata.name }}</bk-button>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('cluster.labels.nameAndId')" v-if="!isClusterMode">
          <template #default="{ row }">
            <div class="flex flex-col py-[6px] h-[50px]">
              <span class="bcs-ellipsis">{{ clusterNameMap[handleGetExtData(row.metadata.uid, 'clusterID')] }}</span>
              <span class="bcs-ellipsis mt-[6px]">{{ handleGetExtData(row.metadata.uid, 'clusterID') }}</span>
            </div>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('k8s.namespace')" prop="metadata.namespace" sortable>
          <template #default="{ row }">
            <bk-button
              class="bcs-button-ellipsis"
              text
              :disabled="isViewEditable"
              @click="goNamespace(row)">
              {{ row.metadata.namespace }}
            </bk-button>
          </template>
        </bk-table-column>
        <bk-table-column label="Reference" min-width="150">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'reference') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Targets" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'targets') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="MinPods" width="100" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'minPods') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="MaxPods" width="100" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'maxPods') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Replicas" width="80" :resizable="false">
          <template #default="{ row }">
            <span>{{ handleGetExtData(row.metadata.uid, 'replicas') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="Age" width="100" sortable="custom" prop="createTime" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <span v-bk-tooltips="{ content: handleGetExtData(row.metadata.uid, 'createTime') }">
              {{ handleGetExtData(row.metadata.uid, 'age') }}</span>
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.source')" :show-overflow-tooltip="false">
          <template #default="{ row }">
            <sourceTableCell
              :row="row"
              :source-type-map="sourceTypeMap" />
          </template>
        </bk-table-column>
        <bk-table-column :label="$t('generic.label.editMode.text')" width="100">
          <template #default="{ row }">
            <span>
              {{handleGetExtData(row.metadata.uid, 'editMode') === 'form'
                ? $t('generic.label.editMode.form') : 'YAML'}}
            </span>
          </template>
        </bk-table-column>
        <bk-table-column
          :label="$t('generic.label.action')"
          :resizable="false"
          width="150"
          fixed="right"
          v-if="!isViewEditable">
          <template #default="{ row }">
            <bk-button
              text
              @click="handleUpdateResource(row)">{{ $t('generic.button.update') }}</bk-button>
            <bk-button
              class="ml10" text
              @click="handleDeleteResource(row)">{{ $t('generic.button.delete') }}</bk-button>
          </template>
        </bk-table-column>
        <template #empty>
          <BcsEmptyTableStatus
            :button-text="$t('generic.button.resetSearch')"
            type="search-empty"
            @clear="handleShowViewConfig" />
        </template>
      </bk-table>
    </template>
    <template #detail="{ data, extData }">
      <HPADetail :data="data" :ext-data="extData"></HPADetail>
    </template>
  </BaseLayout>
</template>
<script>
import { defineComponent } from 'vue';

import sourceTableCell from '../common/source-table-cell.vue';

import HPADetail from './hpa-detail.vue';

import BaseLayout from '@/views/resource-view/common/base-layout';

export default defineComponent({
  name: 'HPA',
  components: { BaseLayout, HPADetail, sourceTableCell },
});
</script>
