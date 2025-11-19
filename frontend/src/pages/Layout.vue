<template>
  <div class="layout">
    <router-view />
    
    <van-tabbar v-model="active" route>
      <van-tabbar-item replace to="/quotes" icon="bar-chart-o">
        行情
      </van-tabbar-item>
      <van-tabbar-item replace to="/trade" icon="exchange">
        交易
      </van-tabbar-item>
      <van-tabbar-item replace to="/positions" icon="records-o">
        持仓
      </van-tabbar-item>
      <van-tabbar-item replace to="/funds" icon="gold-coin-o">
        资金
      </van-tabbar-item>
      <van-tabbar-item replace to="/mine" icon="user-o">
        我的
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const active = ref(0)
</script>

<style scoped>
.layout {
  padding-bottom: 50px;
}
</style>
