<script setup>
import {h, ref} from 'vue';
import {CheckCircleOutlined, FileSearchOutlined, IdcardOutlined, FolderOpenOutlined} from '@ant-design/icons-vue';
import {Comparison, OpenDirectoryDialog} from "../wailsjs/go/main/App";

const open = ref(false);
const pathA = ref('/Volumes/720G-520/pathA');
const pathB = ref('/Volumes/720G-520/pathB');

const checkedA = ref(false);
const checkedB = ref(false);

const loadingName = ref(false)
const disabledName = ref(false)

const loadingMD5= ref(false)
const disabledMD5= ref(false)

function newComparison() {
  open.value = true;
}

function openDirectory(type) {
  OpenDirectoryDialog().then(result => {
    type === 'A' ? pathA.value = result : pathB.value = result
  })
}

function comparison(type) {
  if (type === 'name') {
    loadingName.value = true
    disabledMD5.value = true
  }

  if (type === 'md5') {
    loadingMD5.value = true
    disabledName.value = true
  }

  setTimeout(() => {
    let data = {
      checkType: type,
      pathA: pathA.value,
      isAppendTimeA: checkedA.value,
      pathB: pathB.value,
      isAppendTimeB: checkedB.value,
    }

    Comparison(data).then(result => {
      console.log(result)

      if (type === 'name') {
        loadingName.value = false
        disabledMD5.value = false
      }

      if (type === 'md5') {
        loadingMD5.value = false
        disabledName.value = false
      }
    })
  }, 3000)


}


</script>

<template>
  <img id="empty-comparison" src="./assets/images/empty-comparison.png"/>
  <a-button @click="newComparison" type="primary" size='large' class="comparison-btn" :icon="h(CheckCircleOutlined)">
    对比差异
  </a-button>

  <a-modal v-model:open="open"
           title="配置"
           @ok="handleOk"
           :width="600"
           :footer="null"
           :maskClosable="false"
  >
    <div class="config-item">
      <div class="config-name">
        <span>目录 A</span>
      </div>
      <div class="config-details">
        <div class="config-path">
          <a-input-group compact>
            <a-input v-model:value="pathA" style="width: 360px" placeholder="请设置路径" />
            <a-tooltip title="选择路径">
              <a-button @click="openDirectory('A')">
                <template #icon><FolderOpenOutlined /></template>
              </a-button>
            </a-tooltip>
          </a-input-group>
        </div>
        <div class="append-create-time">
          <a-checkbox v-model:checked="checkedA">文件名追加创建时间</a-checkbox>
        </div>
        <div class="file-num">
          <span>文件数：5987</span>
        </div>
      </div>
    </div>
    <div class="config-item">
      <div class="config-name">
        <span>目录 B</span>
      </div>
      <div class="config-details">
        <div class="config-path">
          <a-input-group compact>
            <a-input v-model:value="pathB" style="width: 360px" placeholder="请设置路径" />
            <a-tooltip title="选择路径">
              <a-button @click="openDirectory('B')">
                <template #icon><FolderOpenOutlined /></template>
              </a-button>
            </a-tooltip>
          </a-input-group>
        </div>
        <div class="append-create-time">
          <a-checkbox v-model:checked="checkedB">文件名追加创建时间</a-checkbox>
        </div>
        <div class="file-num">
          <span>文件数：5987</span>
        </div>
      </div>
    </div>
    <div class="ant-modal-footer">
      <a-button @click="comparison('name')" class="ant-btn ant-btn-default comparison-name" type="button" :icon="h(FileSearchOutlined)" :loading="loadingName" :disabled="disabledName">
        <span>对比文件名</span>
      </a-button>
      <a-button @click="comparison('md5')" class="ant-btn ant-btn-primary" type="button" :icon="h(IdcardOutlined)" :loading="loadingMD5" :disabled="disabledMD5">
        <span>对比MD5</span>
      </a-button>
    </div>
  </a-modal>
</template>

<style lang="scss">
#empty-comparison {
  display: block;
  width: 248px;
  margin: auto;
  padding-top: 180px;
}

.comparison-btn {
  margin-top: 90px;
}

.comparison-name {
  background: #00884e;
  color: white;
}

.comparison-name:not(:disabled):hover {
  color: white;
  border-color: white;
}

.config-item {
  display: flex;
  .config-name {
    width: 20%;
    font-size: 14px;
    line-height: 18px;
    text-align: center;
    font-style: normal;
    text-transform: none;
    padding-top: 9px;
  }
  .config-details {
    width: 80%;
    height: 100px;
    .config-path {
      span {
        text-align: left;
      }
    }

    .append-create-time {
      display: flex;
      justify-content: flex-start;
      margin-top: 10px;
      span {
        font-size: 12px;
        line-height: 22px;
      }
    }

    .file-num {
      display: flex;
      justify-content: flex-start;
      margin-top: 10px;
      font-weight: 400;
      font-size: 12px;
      color: rgba(0,0,0,0.65);
      line-height: 22px;
      text-align: left;
      font-style: normal;
      text-transform: none;
    }
  }
}

.config-item + .config-item {
  margin-top: 30px;
}


</style>
