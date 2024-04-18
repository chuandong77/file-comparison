<script setup>
import {h, reactive, ref} from 'vue';
import { CheckCircleOutlined } from '@ant-design/icons-vue';
const open = ref(false);
//
// const data = reactive({
//   isNewComparison: false,
// })


function newComparison() {
  open.value = true;
}


</script>

<template>
  <img id="empty-comparison" src="./assets/images/empty-comparison.png"/>
  <a-button @click="newComparison" type="primary" :size="size" class="comparison-btn" :icon="h(CheckCircleOutlined)">
    对比差异
  </a-button>

  <a-modal v-model:open="open" title="配置" @ok="handleOk" :width="600">
    <p>Some contents...</p>
    <p>Some contents...</p>
    <p>Some contents...</p>
  </a-modal>
</template>

<style>
#empty-comparison {
  display: block;
  width: 248px;
  margin: auto;
  padding-top: 180px;
}

.comparison-btn {
  margin-top: 90px;
}
</style>
