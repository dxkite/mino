<template>
  <el-form :model="form">
    <el-form-item  v-for="(item, index) in formItem" :key="index">
      <template #label>
        <div class="form-item">
          <div>{{ item.title }}</div>
          <el-tooltip
            v-if="item.description"
            effect="dark"
            :content="item.description"
            placement="top-start"
          >
            <el-icon class="form-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </div>
      </template>
      <el-input
        v-if="item.type == 'string'"
        v-model="form[item.name]"
        autocomplete="off"
        placeholder="请输入"
      />
      <el-radio-group
        v-else-if="item.type == 'boolean'"
        v-model="form[item.name]"
      >
        <el-radio :label="true">开启</el-radio>
        <el-radio :label="false">关闭</el-radio>
      </el-radio-group>

      <el-input-number
        v-else-if="item.type == 'integer'"
        v-model="form[item.name]"
        placeholder="请输入"
      />
    </el-form-item>
  </el-form>
</template>

<script>
export default {
  props: {
    modelValue: Object,
    formItem: Array,
  },
  data() {
    return {
      form: {},
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

<style scoped>
.form-item {
  display: flex;
  align-items: center;
  justify-content: center;
}
.form-icon {
  padding: 0 4px;
}
</style>
