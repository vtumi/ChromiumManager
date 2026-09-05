<template>
  <el-dialog v-model="visible" title="代理列表" :width="800" :before-close="handleClose">
    <div class="proxy-header">
      <el-input
        v-model="keyword"
        placeholder="名称 / IP / 语言 / 时区"
        clearable
        class="proxy-search"
        @keydown.enter="onSearchClick"
        @clear="onSearchClick"
      >
        <template #append>
          <el-button :icon="Search" @click="onSearchClick"></el-button>
        </template>
      </el-input>
      <el-button type="primary" :icon="Plus" @click="onAddClick">添加代理</el-button>
    </div>

    <div class="proxy-table-wrap">
      <el-table
        v-loading="loading"
        :data="list"
        class="proxy-table"
        highlight-current-row
        stripe
        @row-click="onSelectClick"
      >
        <el-table-column
          prop="name"
          label="代理名称"
          width="auto"
          show-overflow-tooltip
        ></el-table-column>
        <el-table-column prop="ip" label="IP" width="120" show-overflow-tooltip></el-table-column>
        <el-table-column
          prop="lang"
          label="语言"
          width="80"
          show-overflow-tooltip
        ></el-table-column>
        <el-table-column
          prop="timezone"
          label="时区"
          width="140"
          show-overflow-tooltip
        ></el-table-column>
        <el-table-column label="位置" width="60" align="center">
          <template #default="scope">
            <el-icon
              v-if="scope.row.location"
              class="location-icon"
              @click.stop="viewMapLocation(scope.row.location)"
            >
              <Location />
            </el-icon>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="150">
          <template #default="scope">
            <el-button size="small" @click.stop="onEditClick(scope.row)">编辑</el-button>
            <el-button size="small" class="delete" @click.stop="onDeleteClick(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无数据"></el-empty>
        </template>
      </el-table>
    </div>

    <el-pagination
      v-model:current-page="page.current"
      v-model:page-size="page.size"
      :total="page.total"
      layout="total, prev, pager, next"
      class="proxy-pagination"
      @current-change="handleCurrentChange"
    ></el-pagination>

    <!-- Form for Add/Edit Proxy -->
    <el-dialog
      v-model="formDialog"
      :title="form._id ? '编辑代理' : '添加代理'"
      :width="650"
      append-to-body
      :before-close="onFormCancel"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="auto">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="代理名称" prop="name">
              <el-input v-model="form.name" placeholder="请输入代理名称"></el-input>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="代理地址" prop="url">
              <el-input v-model="form.url" placeholder="请输入代理地址"></el-input>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="语言">
              <el-select
                v-model="form.lang"
                filterable
                clearable
                placeholder="请选择语言"
                style="width: 100%"
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
                v-model="form.timezone"
                filterable
                clearable
                :fit-input-width="true"
                placeholder="请选择时区"
                style="width: 100%"
              >
                <el-option v-for="tz in timezones" :key="tz" :label="tz" :value="tz"></el-option>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="位置" prop="location">
              <el-input
                v-model="form.location"
                placeholder="请输入或选择位置"
              >
                <template #suffix>
                  <el-icon class="location-pick-icon" @click.stop="onPickLocation">
                    <Location />
                  </el-icon>
                </template>
              </el-input>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="onFormCancel">取消</el-button>
          <el-button type="primary" @click="onFormConfirm">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch, nextTick } from 'vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import { getProxies, getProxy, addProxy, updateProxy, deleteProxy } from '@/api'
import { languages, timezones } from '@/utils/constants'
import { openMapPicker, viewMapLocation } from '@/utils/mapPicker'
import { Location } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'change', 'select'])

const visible = ref(false)
const loading = ref(false)
const list = ref([])
const page = reactive({ current: 1, size: 8, total: 0 })
const keyword = ref('')

const formDialog = ref(false)
const formRef = ref(null)
const form = reactive({})
const rules = {
  name: [{ required: true, message: '请输入代理名称', trigger: 'blur' }],
  url: [
    { required: true, message: '请输入代理地址', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        const pattern = /^(https?|socks[45]?):\/\/([^@]+@)?[^:]+:\d+$/
        if (value && !pattern.test(value)) {
          callback(new Error('代理地址格式不正确'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

watch(
  () => props.modelValue,
  (val) => {
    visible.value = val
    if (val) {
      keyword.value = ''
      page.current = 1
      fetchList()
    }
  }
)

const handleClose = () => {
  emit('update:modelValue', false)
}

const fetchList = async () => {
  loading.value = true
  try {
    const res = await getProxies({
      page: page.current,
      pageSize: page.size,
      keyword: keyword.value
    })
    if (res) {
      list.value = res.list || []
      page.total = res.total || 0
    }
  } catch (err) {
    console.error(err)
  }
  loading.value = false
}

const handleCurrentChange = (val) => {
  page.current = val
  fetchList()
}

const onSearchClick = () => {
  page.current = 1
  fetchList()
}

const onAddClick = () => {
  Object.assign(form, { _id: undefined, name: '', url: '', lang: '', timezone: '', location: '' })
  formDialog.value = true
  nextTick(() => {
    formRef.value && formRef.value.clearValidate()
  })
}

const onEditClick = async (row) => {
  try {
    const data = await getProxy(row._id)
    Object.assign(form, data)
    formDialog.value = true
    nextTick(() => {
      formRef.value && formRef.value.clearValidate()
    })
  } catch (err) {
    if (!err?.silent) ElMessage.error('获取失败: ' + (err?.message || err))
  }
}

const onFormCancel = () => {
  formDialog.value = false
}

const onPickLocation = async () => {
  const loc = await openMapPicker(form.location || '')
  if (loc) form.location = loc
}

const onSelectClick = (row) => {
  emit('select', row._id)
  emit('update:modelValue', false)
}

const onFormConfirm = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (valid) {
      const loadingInstance = ElLoading.service({ fullscreen: true })
      try {
        if (form._id) {
          await updateProxy(form)
          ElMessage.success('编辑成功')
        } else {
          await addProxy(form)
          ElMessage.success('添加成功')
        }
        formDialog.value = false
        fetchList()
        emit('change')
      } catch (err) {
        if (!err?.silent)
          ElMessage.error((form._id ? '编辑失败：' : '添加失败：') + (err?.message || err))
      } finally {
        loadingInstance.close()
      }
    }
  })
}

const onDeleteClick = (row) => {
  ElMessageBox.confirm('是否删除该代理？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  })
    .then(async () => {
      try {
        await deleteProxy(row._id)
        ElMessage.success('删除成功')
        fetchList()
        emit('change')
      } catch (err) {
        if (!err?.silent) ElMessage.error('删除失败')
      }
    })
    .catch(() => {})
}
</script>

<style lang="scss">
.proxy-table {
  .el-table__body-wrapper .el-scrollbar__view {
    height: 100%;
  }
}
</style>

<style lang="scss" scoped>
.proxy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.proxy-search {
  width: 260px;
}
.proxy-table-wrap {
  height: 386px;
}
.proxy-table {
  width: 100%;
  height: 100%;
  padding-top: 15px;
}
.proxy-pagination {
  padding: 15px 0 20px;
  justify-content: center;
}
.delete {
  color: $red-color;
}
</style>
