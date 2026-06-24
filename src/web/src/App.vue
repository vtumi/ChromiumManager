<template>
  <el-config-provider :locale="locale">
    <Left />
    <Right />
    <MapPicker />
  </el-config-provider>
</template>

<script setup>
import { provide, reactive, ref } from 'vue'
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'

import Left from './components/left/LeftIndex.vue'
import Right from './components/right/RightIndex.vue'
import MapPicker from './components/MapPicker.vue'

const locale = zhCn

let activeGroupId = ref(null)
const updateActiveGroupId = (value) => {
  activeGroupId.value = value
}
provide('activeGroupId', activeGroupId)
provide('updateActiveGroupId', updateActiveGroupId)

let device = reactive({
  group: []
})
const updateDeviceGroup = (value) => {
  device.group = value
}
provide('device', device)
provide('updateDeviceGroup', updateDeviceGroup)

const getGroupName = (value) => {
  for (const item of device.group) {
    if (item._id === value) {
      return item.name
    }
  }
}
provide('getGroupName', getGroupName)
</script>

<style>
#app {
  width: 100%;
  height: 100%;
  overflow: hidden;
}
</style>
