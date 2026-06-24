<template>
  <div class="right">
    <div class="header">
      <div class="title">
        <span>配置列表</span>
      </div>
      <div class="form">
        <el-input
          v-model="model.condition.keyword"
          placeholder="名称 / IP / 语言 / 时区"
          clearable
          class="input-with-select"
          @keydown.enter="onSearchClick"
        >
          <template #prepend>
            <el-input
              :model-value="filterProxyName"
              placeholder="全部代理"
              class="input-picker"
              @click="onManageClick"
              @keydown.prevent
            >
              <template #suffix>
                <span
                  class="input-picker-suffix"
                  @click.stop="model.condition.proxyId ? onProxyClear() : onManageClick()"
                >
                  <el-icon v-if="model.condition.proxyId" class="el-input__clear">
                    <CircleClose />
                  </el-icon>
                  <el-icon v-else><ArrowDown /></el-icon>
                </span>
              </template>
            </el-input>
          </template>
        </el-input>
        <el-button
          type="primary"
          size="default"
          :icon="Search"
          class="search"
          @click="onSearchClick"
        ></el-button>
        <el-button
          type="primary"
          size="default"
          :icon="Refresh"
          class="refresh"
          @click="onRefreshClick"
        ></el-button>
        <el-button
          type="primary"
          size="default"
          :icon="Plus"
          class="add"
          @click="onAddClick"
        ></el-button>
      </div>
    </div>
    <Content ref="contentRef" />
    <ProxyManagement v-model="proxyDialog" @change="fetchProxies" @select="onProxySelect" />
  </div>
</template>

<script setup>
import { inject, provide, reactive, ref, computed, watch, onMounted } from 'vue'
import { Search, Refresh, Plus, ArrowDown, CircleClose } from '@element-plus/icons-vue'
import { getProxies } from '@/api'

import Content from './RightContent.vue'
import ProxyManagement from './ProxyManagement.vue'

const activeGroupId = inject('activeGroupId')
let contentRef = ref(null)
let proxyDialog = ref(false)
let proxies = ref([])
let model = reactive({
  condition: { proxyId: '', keyword: '' }
})

const fetchProxies = async () => {
  try {
    const res = await getProxies({ all: 1 })
    if (res) proxies.value = res
  } catch (e) {
    console.error(e)
  }
}

provide('proxies', proxies)
provide('fetchProxies', fetchProxies)

onMounted(() => {
  fetchProxies()
})

watch(activeGroupId, () => {
  model.condition = { proxyId: '', keyword: '' }
})

const filterProxyName = computed(() => {
  const p = proxies.value.find((item) => item._id === model.condition.proxyId)
  return p ? p.name : ''
})

const onManageClick = () => {
  proxyDialog.value = true
}

const onProxySelect = async (proxyId) => {
  await fetchProxies()
  model.condition.proxyId = proxyId
  onSearchClick()
}

const onProxyClear = () => {
  model.condition.proxyId = ''
  onSearchClick()
}

const onSearchClick = () => {
  contentRef.value.onSearchClick(model.condition.proxyId, model.condition.keyword)
}
const onRefreshClick = () => {
  contentRef.value.onRefreshClick()
}
const onAddClick = () => {
  contentRef.value.onAddClick()
}
</script>

<style lang="scss">
.right .header .form {
  .el-input-group__prepend .el-input {
    width: 120px;

    .el-input__wrapper {
      box-shadow: none;
      background-color: transparent;

      .el-input__inner {
        height: 36px;
        line-height: 36px;
        text-indent: 10px;
      }
    }
  }

  .input-with-select .el-input-group__prepend {
    background-color: $white-color;
    padding: 0;
  }

  .input-with-select .el-input__wrapper {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
  }

  .el-button {
    height: 38px;
    line-height: 38px;

    .el-icon {
      font-size: 20px;
    }
  }
}
</style>

<style lang="scss" scoped>
.right {
  margin-left: 280px;
  height: 100%;

  .header {
    padding: 0 20px;
    height: 60px;
    line-height: 60px;
    overflow: hidden;
    border-bottom: $border1;

    .title {
      float: left;
      font-size: 24px;
      color: $text-color1;
    }

    .form {
      float: right;
      display: flex;
      padding: 10px 0;
      width: 50%;
      min-width: 510px;

      .search {
        border-top-left-radius: 0;
        border-bottom-left-radius: 0;
      }
    }
  }
}
</style>
