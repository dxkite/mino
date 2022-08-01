<template>
  <div class="table">
    <el-table :data="tableData" style="width: 100%" row-key="id" border>
      <el-table-column label="远程客户端" prop="src" width="200" />
      <el-table-column label="协议" prop="protocol" width="80" />
      <el-table-column label="目标网站">
        <template #default="scope">
          <div :title="scope.row.dst" class="table-dst">
            {{ scope.row.dst }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="上传流量" width="100">
        <template #default="scope">
          {{ bytesToSize(scope.row.up) }}
        </template>
      </el-table-column>
      <el-table-column label="下载流量" width="100">
        <template #default="scope">
          {{ bytesToSize(scope.row.down) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="scope">
          <el-button
            v-if="!scope.row.isGroup"
            size="small"
            type="danger"
            @click="handleDisconnect(scope.row)"
            >断开连接
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
export default {
  name: "table",
  props: {
    tableData: Array,
  },
  data() {
    return {};
  },
  methods: {
    // 单位转换
    bytesToSize(bytes) {
      if (bytes === 0) return "0 B";
      var k = 1000, // or 1024
        sizes = ["B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"],
        i = Math.floor(Math.log(bytes) / Math.log(k));

      return (bytes / Math.pow(k, i)).toPrecision(3) + " " + sizes[i];
    },
    // 断开连接
    handleDisconnect(row) {
      this.$emit('handleDisconnect', row)
    },
  },
};
</script>

<style scoped>
.table-dst {
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}
</style>