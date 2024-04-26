<script setup>
import {computed, h, onMounted, reactive, ref, watch} from 'vue';
import {CheckCircleOutlined, DeleteOutlined, FolderOpenOutlined, QuestionCircleOutlined} from '@ant-design/icons-vue';
import {DelComparisonResult, GetComparisonResult, OpenFileDialog, DelFile} from "../../wailsjs/go/main/App";

import {useRouter} from 'vue-router'
import {message} from "ant-design-vue";

const router = useRouter()
const isCompareAgain = ref(false);

const loadingDelA = ref(false)
const loadingDelB = ref(false)
const disabledDelA = ref(false)
const disabledDelB = ref(false)
const spinning = ref(false)
const data = ref([])

const pageSize = ref(10);
const current = ref(1);
const total = ref(0);

const columns = [
  {
    title: 'A 目录文件',
    dataIndex: 'fileNameA',
  },
  {
    title: 'B 目录文件',
    dataIndex: 'fileNameB',
  },
  {
    title: '操作',
    dataIndex: 'key',
  },
];

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

const params = ref([]);
const onSelectChange = selectedRowKeys => {
  params.value = []
  for (var i = 0; i < selectedRowKeys.length; i ++) {
    let row = data.value[selectedRowKeys[i]]
    let param = {
      PathA: row.pathA + '/' + row.fileNameA,
      PathB: row.pathB + '/' + row.fileNameB,
      Index: (current.value - 1) * pageSize.value + selectedRowKeys[i]
    }
    params.value.push(param)
  }

  state.selectedRowKeys = selectedRowKeys;
};

watch(current, () => {
  getResult(current.value, pageSize.value)
});

onMounted(() => {
    getResult(1, pageSize.value)
})

function getResult(page, pageSize) {
  //清空已选择的文件
  params.value = []
  state.selectedRowKeys = []

  //重置页码
  current.value = page

  spinning.value = true
  GetComparisonResult(page, pageSize).then(result => {
    data.value = []
    let res = JSON.parse(result)

    total.value = res.data.Total
    if (res.ret === 1 && res.data.List !== null) {
      for (var i = 0; i < res.data.List.length; i++) {
        var item = res.data.List[i];
        data.value.push({
          key: i,
          pathA: item.PathA,
          fileNameA: item.NameA,
          fileSizeA: formatBytes(item.SizeA),
          base64A: item.Base64A,

          pathB: item.PathB,
          fileNameB: item.NameB,
          fileSizeB: formatBytes(item.SizeB),
          base64B: item.Base64B,

          isSameSize: item.SizeA === item.SizeB
        });
      }
    }

    spinning.value = false
  })
}

function compareAgain() {
  isCompareAgain.value = true
}

function compareAgainOK() {
  DelComparisonResult().then(result => {
    router.push('/home')
  })
}

function viewSourceFile(pathA, pathB) {
  OpenFileDialog(pathA, pathB).then(result => {
  })
}

//修改分页大小
const onShowSizeChange = (current, pageSize) => {
  getResult(1, pageSize)
};

function formatBytes(bytes) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function delFile(type) {
  if (params.value.length <= 0) {
    message.warning('请选择要删除的文件');
    return
  }

  DelFile(params.value, type).then(result => {
    let res = JSON.parse(result)
    if (res.ret === 0) {
      message.error(res.msg);
      return
    }

    message.info('删除成功');

    getResult(1, pageSize.value)
  })
}

</script>

<template>
  <div class="container">
    <a-spin class="c-spin" size="large" tip="正在生成预览..." :spinning="spinning"/>
    <div class="head">
      <a-popconfirm title="删除后将不可恢复，确定删除？" @confirm="delFile('A')" ok-text="确定" cancel-text="取消">
        <template #icon><question-circle-outlined style="color: red" /></template>
        <a-button class="del-a" danger type="primary" :icon="h(DeleteOutlined)" :loading="loadingDelA" :disabled="disabledDelA">
          <span>删除 A 中数据</span>
        </a-button>
      </a-popconfirm>

      <a-popconfirm title="删除后将不可恢复，确定删除？" @confirm="delFile('B')" ok-text="确定" cancel-text="取消">
        <template #icon><question-circle-outlined style="color: red" /></template>
        <a-button class="del-b" danger :icon="h(DeleteOutlined)" :loading="loadingDelB" :disabled="disabledDelB">
          <span>删除 B 中数据</span>
        </a-button>
      </a-popconfirm>

      <a-button @click="compareAgain()" size='large' class="compare-again" type="primary" :icon="h(CheckCircleOutlined)">
        <span>重新对比</span>
      </a-button>
    </div>
    <div class="content">
      <a-table
          :row-selection="{ selectedRowKeys: state.selectedRowKeys, onChange: onSelectChange }"
          :columns="columns"
          :data-source="data"
          class="table"
          :pagination="false"
      >
        <template #emptyText>
          <span>暂无重复数据</span>
        </template>
        <template #bodyCell="{column, text, record}" >
          <template v-if="column.dataIndex === 'fileNameA'">
            <div class="file-img">
              <a-image
                       :width="60"
                       :src="'data:image/*;base64,' + record.base64A"
                       fallback="data:image/png;base64,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"

              />
            </div>
            <div class="file-info">
              <span class="file-name">{{ text }}</span><br/>
              <span  class="file-size">{{ record.fileSizeA }}</span>
            </div>
          </template>

          <template v-if="column.dataIndex === 'fileNameB'">
            <div class="file-img">
              <a-image
                  :width="60"
                  :src="'data:image/*;base64,' + record.base64B"
                  fallback="data:image/png;base64,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"

              />
            </div>
            <div class="file-info">
              <span class="file-name">{{ text }}</span><br/>
              <span class="file-size" :style="{color: record.isSameSize ? '#767676' : 'red' }">{{ record.fileSizeB }}</span>
            </div>

          </template>

          <template v-if="column.dataIndex === 'key'">
            <a-button @click="viewSourceFile(record.pathA + '/' +record.fileNameA, record.pathB + '/' +record.fileNameB)" class="del-a" type="link" :icon="h(FolderOpenOutlined)">
              <span>查看源文件</span>
            </a-button>
          </template>

        </template >
      </a-table>
      <a-pagination class="tab-pagination"
                    v-model:current="current"
                    v-model:pageSize="pageSize"
                    show-size-changer
                    v-model:total="total"
                    @showSizeChange="onShowSizeChange"
      >
        <template #buildOptionText="props">
          <span>{{ props.value }}条/页</span>
        </template>
      </a-pagination>
    </div>
  </div>
  <a-modal v-model:open="isCompareAgain" title="提示" @ok="compareAgainOK" ok-text="确定" cancel-text="取消" width="400px">
    <p style="margin-top: 20px">重新对比将会删除当前的对比结果，确定重新对比？</p>
  </a-modal>
</template>

<style lang="scss">
  .container {
    .c-spin {
      position: fixed;
      top: 50vh;
      z-index: 100;
      left: 46vw;
    }
    .head {
      height: 80px;
      width: 100%;
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: fixed;
      top: 0;
      z-index: 100;
      background: white;

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

    .tab-pagination {
      margin-top: 20px;
      padding-bottom: 20px;
    }

    .content {
      margin-top: 80px;

      .file-info {
        float: left;
        width: 300px;
        height: 60px;
        margin-left: 12px;
        .file-name {
          font-weight: 400;
          font-size: 14px;
          color: #3D3D3D;
          text-align: left;
          font-style: normal;
          text-transform: none;
          line-height: 30px;
        }
        .file-size {
          font-weight: 400;
          font-size: 14px;
          color: #767676;
          line-height: 30px;
          text-align: left;
          font-style: normal;
          text-transform: none;
        }
      }

      .file-img {
        float: left;
      }
    }
  }
</style>