<script setup>
import {computed, h, reactive, ref} from 'vue';
import {CheckCircleOutlined, DeleteOutlined} from '@ant-design/icons-vue';
import {DelComparisonResult} from "../../wailsjs/go/main/App";

import {useRouter} from 'vue-router'

const router = useRouter()
const isCompareAgain = ref(false);

const columns = [
  {
    title: 'A 目录文件',
    dataIndex: 'fileNameA',
    slots: {
      customRender: 'makeFileInfo'
    }
  },
  {
    title: 'B 目录文件',
    dataIndex: 'fileNameB',
    slots: {
      customRender: 'makeFileInfo'
    }
  },
  {
    title: '操作',
    dataIndex: 'key',
  },
];
const data = [];
for (let i = 0; i < 46; i++) {
  data.push({
    key: i,
    pathA: '',
    fileNameA: '测试点A.jpg',
    fileSizeA: 122121,
    pathB: '',
    fileNameB: '测试点B.jpg',
    fileSizeB: 54545454,
  });
}
const state = reactive({
  selectedRowKeys: [],
  // Check here to configure the default column
  loading: false,
});
const hasSelected = computed(() => state.selectedRowKeys.length > 0);
const start = () => {
  state.loading = true;
  // ajax request after empty completing
  setTimeout(() => {
    state.loading = false;
    state.selectedRowKeys = [];
  }, 1000);
};
const onSelectChange = selectedRowKeys => {
  console.log('selectedRowKeys changed: ', selectedRowKeys);
  state.selectedRowKeys = selectedRowKeys;
};

function compareAgain() {
  isCompareAgain.value = true
}

function compareAgainOK() {
  DelComparisonResult().then(result => {
    router.push('/home')
  })
}

</script>

<template>
  <div class="container">
    <div class="head">
      <a-button @click="comparison('name')" class="del-a" danger type="primary" :icon="h(DeleteOutlined)" :loading="loadingDelA" :disabled="disabledDelA">
        <span>删除 A 中数据</span>
      </a-button>
      <a-button @click="comparison('name')" class="del-b" danger :icon="h(DeleteOutlined)" :loading="loadingDelB" :disabled="disabledDelB">
        <span>删除 B 中数据</span>
      </a-button>
      <a-button @click="compareAgain()" size='large' class="compare-again" type="primary" :icon="h(CheckCircleOutlined)">
        <span>重新对比</span>
      </a-button>
    </div>
    <div class="head-placeholder"></div>
    <div class="content">
      <a-table
          :row-selection="{ selectedRowKeys: state.selectedRowKeys, onChange: onSelectChange }"
          :columns="columns"
          :data-source="data"
          class="table"
      >
        <!--创建文件详情插槽-->
        <template #makeFileInfo="{text, record, index}" >
          <span style="color: red">{{ text }}</span>
        </template >
      </a-table>
    </div>
  </div>
  <a-modal v-model:open="isCompareAgain" title="提示" @ok="compareAgainOK" ok-text="确定" cancel-text="取消" width="400px">
    <p style="margin-top: 20px">重新对比将会删除当前的对比结果，确定重新对比？</p>
  </a-modal>
</template>

<style lang="scss">
  .container {
    .head {
      height: 80px;
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .del-a {
        margin-left: 28px;
        margin-right: 14px;
      }

      .del-b {
        margin-right: auto;
      }

      .compare-again {
        margin-right: 28px;
      }
    }

    .table {
      margin-left: 28px;
      margin-right: 28px;
    }
  }
</style>