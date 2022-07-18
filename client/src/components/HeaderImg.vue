<template>
  <div>
    <img class="header-img" src="@/assets/logo.svg" />
    <p class="header-title">
      <slot name="title">MINO 网络访问助手</slot>
    </p>
  </div>
</template>

<script>
export default {};
</script>

<style scoped>
.header-img {
  display: flex;
  flex-direction: column;

  width: 476px;
  height: 143.64px;
}
.header-title {
  font-family: PingFang SC;
  font-size: 48px;
  font-weight: 400;
  line-height: 67px;
  letter-spacing: 0em;
  text-align: left;
}
</style>
