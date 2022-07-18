<template>
  <el-form :model="form">
    <el-form-item label="监听地址" :label-width="formLabelWidth">
      <el-input
        v-model="form.address"
        autocomplete="off"
        placeholder="请输入"
      />
    </el-form-item>
    <el-form-item label="自动环回地址检测">
      <el-radio-group v-model="form.host_detect_loopback">
        <el-radio :label="true">开启</el-radio>
        <el-radio :label="false">关闭</el-radio>
      </el-radio-group>
    </el-form-item>
    <el-form-item label="hot_load">
      <el-input
        type="number"
        v-model="form.hot_load"
        autocomplete="off"
        placeholder="请输入"
      />
    </el-form-item>
  </el-form>
</template>

<script>
export default {
  props: {
    modelValue: Object,
  },
  data() {
    return {
      form: {
        address: "",
        host_detect_loopback: false,
        hot_load: 10000,
      },
    };
  },
  watch: {
    form() {
      this.$emit("update:modelValue", this.form);
    },
  },
  mounted() {
    console.log("mounted");
    this.form = this.modelValue;
  },
};
</script>

<style></style>
