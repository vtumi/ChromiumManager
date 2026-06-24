<template>
  <div class="content">
    <div ref="listRef" class="list">
      <el-table
        ref="tableRef"
        :data="listView"
        highlight-current-row
        stripe
        class="table"
        @row-click="onItemClick"
      >
        <el-table-column prop="name" label="名称" width="auto" class-name="left"></el-table-column>
        <el-table-column prop="proxyName" label="代理" width="auto"></el-table-column>
        <el-table-column prop="ip" label="IP" width="120"></el-table-column>
        <el-table-column prop="lang" label="语言" width="80"></el-table-column>
        <el-table-column prop="timezone" label="时区" width="140"></el-table-column>
        <el-table-column label="位置" width="60" align="center">
          <template #default="scope">
            <el-icon
              v-if="scope.row.location"
              class="location-icon"
              @click="viewMapLocation(scope.row.location)"
            >
              <Location />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column label="Cookie" width="80" align="center">
          <template #default="scope">
            <el-icon
              :class="['cookie-icon', { disabled: runningSet.has(scope.row._id) }]"
              @click.stop="!runningSet.has(scope.row._id) && onCookieImport(scope.row)"
            >
              <Upload />
            </el-icon>
            <el-icon
              :class="['cookie-icon', { disabled: runningSet.has(scope.row._id) }]"
              @click.stop="!runningSet.has(scope.row._id) && onCookieExport(scope.row)"
            >
              <Download />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="200">
          <template #default="scope">
            <template v-if="runningSet.has(scope.row._id)">
              <el-button type="primary" size="small" @click.stop="onShowClick(scope.row)">
                激活
              </el-button>
              <el-button type="danger" size="small" @click.stop="onStopClick(scope.row)">
                关闭
              </el-button>
              <el-button size="small" @click.stop="onEditClick(scope.row)">编辑</el-button>
            </template>
            <template v-else>
              <el-button type="primary" size="small" @click.stop="onLaunchClick(scope.row)">
                启动
              </el-button>
              <el-button size="small" @click.stop="onEditClick(scope.row)">编辑</el-button>
              <el-dropdown trigger="click" @command="(cmd) => onMoreCommand(cmd, scope.row)">
                <el-button size="small" @click.stop>更多</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="delete" class="delete-item">删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无数据"></el-empty>
        </template>
      </el-table>
    </div>
    <div class="pagination">
      <el-pagination
        v-model:current-page="model.page.current"
        layout="total, prev, pager, next"
        :page-size="model.page.size"
        :total="model.page.total"
        @current-change="handleCurrentChange"
      ></el-pagination>
    </div>
    <el-dialog
      v-model="formDialog"
      :title="model.form._id ? '编辑配置' : '添加配置'"
      :width="800"
      :before-close="onFormCancel"
    >
      <el-form
        ref="formRef"
        :model="model.form"
        :rules="model.rules"
        label-position="right"
        label-width="auto"
        size="default"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="名称" prop="name">
              <el-input v-model="model.form.name" placeholder="请输入名称"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分组" prop="groupId">
              <el-select v-model="model.form.groupId" filterable>
                <template v-for="item in model.group" :key="item._id">
                  <el-option :label="item.name" :value="item._id"></el-option>
                </template>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="代理" prop="proxy" class="input-picker">
              <el-input
                :model-value="formProxyName"
                placeholder="请选择代理"
                @click="onManageProxyClick"
                @keydown.prevent
              >
                <template #suffix>
                  <span
                    class="input-picker-suffix"
                    @click.stop="model.form.proxy ? (model.form.proxy = '') : onManageProxyClick()"
                  >
                    <el-icon v-if="model.form.proxy" class="el-input__clear">
                      <CircleClose />
                    </el-icon>
                    <el-icon v-else><ArrowDown /></el-icon>
                  </span>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排序" prop="sort">
              <el-input-number
                v-model="model.form.sort"
                :min="0"
                :max="99"
                controls-position="right"
              ></el-input-number>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row class="switches-row">
          <el-col>
            <el-form-item label="随机指纹">
              <el-switch v-model="model.form.fp.randomFingerprint" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理语言">
              <el-switch v-model="model.form.fp.proxyLang" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理时区">
              <el-switch v-model="model.form.fp.proxyTimezone" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="代理位置">
              <el-switch v-model="model.form.fp.proxyLocation" />
            </el-form-item>
          </el-col>
          <el-col>
            <el-form-item label="WebRTC">
              <el-switch
                :model-value="!model.form.fp.disableFeatures.includes('webrtc')"
                @update:model-value="(v) => toggleFeature('webrtc', !v)"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="操作系统">
              <el-select v-model="model.form.fp.platform" clearable placeholder="请选择操作系统">
                <el-option label="Windows" value="windows"></el-option>
                <el-option label="Linux" value="linux"></el-option>
                <el-option label="macOS" value="macos"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="浏览器品牌">
              <el-select v-model="model.form.fp.brand" clearable placeholder="请选择浏览器品牌">
                <el-option label="Chrome" value="Chrome"></el-option>
                <el-option label="Edge" value="Edge"></el-option>
                <el-option label="Opera" value="Opera"></el-option>
                <el-option label="Vivaldi" value="Vivaldi"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="设备核心">
              <el-select
                v-model="model.form.fp.hardwareConcurrency"
                placeholder="请选择设备核心"
                clearable
              >
                <el-option label="2" value="2"></el-option>
                <el-option label="4" value="4"></el-option>
                <el-option label="6" value="6"></el-option>
                <el-option label="8" value="8"></el-option>
                <el-option label="10" value="10"></el-option>
                <el-option label="12" value="12"></el-option>
                <el-option label="16" value="16"></el-option>
                <el-option label="20" value="20"></el-option>
                <el-option label="24" value="24"></el-option>
                <el-option label="32" value="32"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="设备内存">
              <el-select
                v-model="model.form.fp.deviceMemory"
                placeholder="请选择设备内存"
                clearable
              >
                <el-option label="2" value="2"></el-option>
                <el-option label="4" value="4"></el-option>
                <el-option label="8" value="8"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="屏幕尺寸">
              <el-select v-model="model.form.fp.screen" clearable placeholder="请选择屏幕尺寸">
                <el-option v-for="s in screens" :key="s" :label="s" :value="s"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="位置" class="input-picker">
              <el-input
                v-model="model.form.fp.location"
                :disabled="model.form.fp.proxyLocation"
                placeholder="请选择位置"
                clearable
                class="input-picker"
                @click="onPickFpLocation"
                @keydown.prevent
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="语言">
              <el-select
                v-model="model.form.fp.lang"
                filterable
                clearable
                :disabled="model.form.fp.proxyLang"
                placeholder="请选择语言"
              >
                <el-option
                  v-for="lang in languages"
                  :key="lang"
                  :label="lang"
                  :value="lang"
                ></el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="时区">
              <el-select
                v-model="model.form.fp.timezone"
                filterable
                clearable
                :disabled="model.form.fp.proxyTimezone"
                :fit-input-width="true"
                placeholder="请选择时区"
              >
                <el-option v-for="tz in timezones" :key="tz" :label="tz" :value="tz"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="禁用伪装">
              <el-checkbox-group
                v-model="model.form.fp.disableFingerprint"
                :disabled="!model.form.fp.randomFingerprint"
              >
                <el-checkbox value="font">字体</el-checkbox>
                <el-checkbox value="audio">音频</el-checkbox>
                <el-checkbox value="canvas">Canvas</el-checkbox>
                <el-checkbox value="clientrects">ClientRects</el-checkbox>
                <el-checkbox value="webgl">WebGL</el-checkbox>
                <el-checkbox value="gpu">GPU</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="额外参数" prop="args">
              <el-input
                v-model="model.form.args"
                placeholder="请输入额外参数，多个以空格分隔"
                type="textarea"
                :rows="3"
                resize="none"
              ></el-input>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button size="default" @click="onFormCancel">取消</el-button>
          <el-button type="primary" size="default" @click="onFormConfirm">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <ProxyManagement v-model="proxyManageVisible" @change="fetchProxies" @select="onProxySelect" />

    <el-dialog v-model="cookieDialog" title="导入Cookie" :width="600">
      <el-form ref="cookieFormRef" :model="cookieForm" :rules="cookieRules">
        <el-form-item prop="text">
          <el-input
            v-model="cookieForm.text"
            placeholder="请输入Cookie，JSON格式字符串"
            type="textarea"
            :rows="10"
            resize="none"
          ></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cookieDialog = false">取消</el-button>
        <el-button type="primary" @click="onCookieImportConfirm">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import {
  computed,
  inject,
  nextTick,
  reactive,
  ref,
  onMounted,
  onUnmounted,
  toRef,
  watch
} from 'vue'
import { Location, ArrowDown, CircleClose, Upload, Download } from '@element-plus/icons-vue'
import { viewMapLocation, openMapPicker } from '@/utils/mapPicker'

import { ElMessage, ElMessageBox } from 'element-plus'

import {
  getProfiles,
  getProfile,
  addProfile,
  updateProfile,
  deleteProfile,
  launchProfile,
  stopProfile,
  showProfile,
  exportCookies,
  importCookies
} from '@/api'
import { validateForm } from '@/utils/common'
import { languages, timezones, screens, BASE_URL } from '@/utils/constants'
import ProxyManagement from './ProxyManagement.vue'

const activeGroupId = inject('activeGroupId')
const device = inject('device')
const proxies = inject('proxies')
const fetchProxies = inject('fetchProxies')

const listRef = ref(null)
const tableRef = ref(null)
const formRef = ref(null)
const formDialog = ref(false)
const runningSet = ref(new Set())
const proxyManageVisible = ref(false)
const selectedRow = ref(null)

const cookieDialog = ref(false)
const cookieImportRow = ref(null)
const cookieFormRef = ref(null)
const cookieForm = reactive({ text: '' })
const cookieRules = {
  text: [
    {
      required: true,
      message: '请输入Cookie',
      trigger: 'blur'
    },
    {
      validator: (rule, value, callback) => {
        let arr
        try {
          arr = JSON.parse(value)
        } catch {}
        if (!Array.isArray(arr) || arr.length === 0) return callback(new Error('Cookie格式不正确'))
        callback()
      },
      trigger: 'blur'
    }
  ]
}

let model = reactive({
  group: toRef(device, 'group'),
  condition: { groupId: activeGroupId, proxyId: '', keyword: '' },

  list: [],
  page: { current: 1, size: 0, total: 0 },
  form: {},
  rules: {
    name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
    groupId: [{ required: true, message: '请选择分组', trigger: 'change' }],
    sort: [{ required: true, message: '请输入排序', trigger: 'blur' }]
  }
})

let resizeTimeout
const resize = () => {
  clearTimeout(resizeTimeout)
  resizeTimeout = setTimeout(() => {
    model.page.size = getPageSize()
    search()
  }, 200)
}

let runningEventSource = null

onMounted(() => {
  window.addEventListener('resize', resize)
  runningEventSource = new EventSource(`${BASE_URL}/events`)
  runningEventSource.onmessage = (e) => {
    try {
      const ids = JSON.parse(e.data)
      if (Array.isArray(ids)) runningSet.value = new Set(ids)
    } catch {}
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', resize)
  if (runningEventSource) runningEventSource.close()
})

watch(activeGroupId, (newVal) => {
  model.condition.groupId = newVal
  model.condition.proxyId = ''
  model.condition.keyword = ''
  model.page.current = 1
  selectedRow.value = null
  nextTick(() => {
    tableRef.value?.setCurrentRow(null)
  })
  resize()
})

const getPageSize = () => {
  return Math.floor((listRef.value.offsetHeight - 40) / 41.38)
}

const proxyMap = computed(() => new Map(proxies.value.map((p) => [p._id, p])))
const formProxyName = computed(() => {
  const p = proxyMap.value.get(model.form.proxy)
  return p ? p.name : ''
})

const listView = computed(() => {
  return model.list.map((row) => {
    let fp = {}
    fp = row.fingerprint || {}
    const proxy = proxyMap.value.get(row.proxy)
    return {
      ...row,
      proxyName: proxy ? proxy.name : row.proxy || '',
      ip: proxy ? proxy.ip || fp.ip || '' : fp.ip || '',
      lang: proxy && fp.proxyLang ? proxy.lang || fp.lang || '' : fp.lang || '',
      timezone: proxy && fp.proxyTimezone ? proxy.timezone || fp.timezone || '' : fp.timezone || '',
      location: proxy ? proxy.location || '' : ''
    }
  })
})

const search = async () => {
  try {
    const params = {
      groupId: model.condition.groupId,
      proxyId: model.condition.proxyId,
      keyword: model.condition.keyword,
      page: model.page.current,
      pageSize: model.page.size
    }
    const res = await getProfiles(params)
    if (res) {
      model.list = res.list || []
      model.page.total = res.total || 0
      selectedRow.value = null
      nextTick(() => {
        tableRef.value?.setCurrentRow(null)
      })
    }
  } catch (err) {
    console.error(err)
  }
}

const handleCurrentChange = (value) => {
  model.page.current = value
  search()
}

const onSearchClick = (proxyId, keyword) => {
  model.condition.proxyId = proxyId
  model.condition.keyword = keyword
  model.page.current = 1
  search()
}

const onRefreshClick = () => {
  search()
}

const defaultFp = () => ({
  platform: '',
  brand: '',
  hardwareConcurrency: '',
  deviceMemory: '',
  disableFeatures: ['webrtc'],
  screen: '',
  lang: '',
  timezone: '',
  location: '',
  disableFingerprint: [],
  randomFingerprint: true,
  proxyLang: true,
  proxyTimezone: true,
  proxyLocation: true
})

const buildPayload = () => {
  const { fp, ...rest } = model.form
  return { ...rest, fingerprint: fp }
}

const onItemClick = (row) => {
  if (selectedRow.value && selectedRow.value._id === row._id) {
    selectedRow.value = null
    nextTick(() => {
      tableRef.value?.setCurrentRow(null)
    })
  } else {
    selectedRow.value = row
  }
}

const onAddClick = () => {
  if (selectedRow.value) {
    const fpObj = selectedRow.value.fingerprint
      ? JSON.parse(JSON.stringify(selectedRow.value.fingerprint))
      : defaultFp()
    model.form = {
      name: '',
      groupId: selectedRow.value.groupId || activeGroupId.value,
      sort: selectedRow.value.sort || 0,
      proxy: selectedRow.value.proxy || '',
      args: selectedRow.value.args || '',
      notes: selectedRow.value.notes || '',
      fp: fpObj
    }
  } else {
    model.form = {
      name: '',
      groupId: activeGroupId.value,
      sort: 0,
      proxy: '',
      args: '',
      notes: '',
      fp: defaultFp()
    }
  }
  nextTick(() => {
    formRef.value && formRef.value.clearValidate()
    formDialog.value = true
  })
}

const onEditClick = async (item) => {
  try {
    const params = await getProfile(item._id)
    let fp = defaultFp()
    if (params.fingerprint && typeof params.fingerprint === 'object') {
      Object.assign(fp, params.fingerprint)
    }
    model.form = { ...params, fp }
    if (!model.form.groupId) model.form.groupId = 'all'
    nextTick(() => {
      formRef.value && formRef.value.clearValidate()
      formDialog.value = true
    })
  } catch (err) {
    if (!err?.silent) ElMessage.error('获取失败: ' + (err?.message || err))
  }
}

const onFormCancel = () => {
  formDialog.value = false
}

const onFormConfirm = async () => {
  let ret = await validateForm(formRef.value)
  if (ret) {
    try {
      if (model.form._id) {
        await updateProfile(buildPayload())
        ElMessage({ type: 'success', showClose: true, message: '编辑成功！' })
      } else {
        await addProfile(buildPayload())
        ElMessage({ type: 'success', showClose: true, message: '添加成功！' })
      }
      formDialog.value = false
      search()
    } catch (err) {
      if (!err?.silent)
        ElMessage({
          type: 'error',
          showClose: true,
          message: (model.form._id ? '编辑失败：' : '添加失败：') + (err?.message || err)
        })
    }
  }
}

const onDeleteClick = (item) => {
  ElMessageBox.confirm('是否删除该配置？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        await deleteProfile(item._id)
        ElMessage({
          type: 'success',
          showClose: true,
          message: '删除成功！'
        })
        search()
      } catch (err) {
        if (!err?.silent) ElMessage({ type: 'error', showClose: true, message: '删除失败！' })
      }
    })
    .catch(() => {})
}

const onLaunchClick = async (row) => {
  try {
    await launchProfile({ id: row._id })
    runningSet.value = new Set([...runningSet.value, row._id])
    ElMessage.success('启动成功！')
  } catch (err) {
    if (!err?.silent) ElMessage.error('启动失败: ' + (err?.message || err))
  }
}

const onShowClick = async (row) => {
  try {
    await showProfile(row._id)
  } catch (err) {
    if (!err?.silent) ElMessage.error('激活失败: ' + (err?.message || err))
  }
}

const onStopClick = async (row) => {
  try {
    await stopProfile(row._id)
    const next = new Set(runningSet.value)
    next.delete(row._id)
    runningSet.value = next
    ElMessage.success('关闭成功！')
  } catch (err) {
    if (!err?.silent) ElMessage.error('关闭失败: ' + (err?.message || err))
  }
}

const onMoreCommand = (cmd, row) => {
  if (cmd === 'delete') onDeleteClick(row)
}

const onCookieImport = (row) => {
  cookieImportRow.value = row
  cookieForm.text = ''
  cookieDialog.value = true
  nextTick(() => cookieFormRef.value?.clearValidate())
}

const onCookieImportConfirm = async () => {
  if (!(await validateForm(cookieFormRef.value))) return
  const cookies = JSON.parse(cookieForm.text)
  try {
    await importCookies({ id: cookieImportRow.value._id, cookies })
    ElMessage.success('导入成功！')
    cookieDialog.value = false
  } catch (err) {
    if (!err?.silent) ElMessage.error('导入失败: ' + (err?.message || err))
  }
}

const onCookieExport = async (row) => {
  try {
    const cookies = await exportCookies(row._id)
    const text = JSON.stringify(cookies, null, 2)
    await navigator.clipboard.writeText(text)
    ElMessage.success('已导出到剪贴板！')
  } catch (err) {
    if (!err?.silent) ElMessage.error('导出失败: ' + (err?.message || err))
  }
}

const toggleFeature = (feature, enabled) => {
  const arr = model.form.fp.disableFeatures
  if (enabled) {
    if (!arr.includes(feature)) arr.push(feature)
  } else {
    const idx = arr.indexOf(feature)
    if (idx > -1) arr.splice(idx, 1)
  }
}

const onPickFpLocation = async () => {
  if (model.form.fp.proxyLocation) return
  const loc = await openMapPicker(model.form.fp.location || '')
  if (loc) model.form.fp.location = loc
}

const onManageProxyClick = () => {
  proxyManageVisible.value = true
}

const onProxySelect = async (proxyId) => {
  await fetchProxies()
  model.form.proxy = proxyId
}

defineExpose({
  onSearchClick,
  onRefreshClick,
  onAddClick
})
</script>

<style lang="scss">
.right .content .list .table {
  .el-table__body-wrapper .el-scrollbar__view {
    height: 100%;
  }

  .cell {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .cookie-icon {
    cursor: pointer;
    font-size: 16px;
    color: #409eff;
    margin: 0 4px;
    &.disabled {
      opacity: 0.3;
      cursor: not-allowed;
    }
  }

  .left .cell {
    padding-left: 24px;
  }
}

.delete-item {
  color: #f56c6c !important;
}
</style>

<style lang="scss" scoped>
.content {
  width: 100%;
  height: calc(100% - 60px);

  .list {
    width: 100%;
    height: calc(100% - 52px);
    overflow: hidden;

    .table {
      height: 100%;

      .delete {
        color: $red-color;
      }
    }
  }

  .pagination {
    padding: 10px;

    .el-pagination {
      justify-content: center;
    }
  }
}

.el-select {
  width: 100%;
}

.switches-row {
  :deep(.el-col) {
    flex: 1;
  }
}
</style>
