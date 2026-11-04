<script setup lang="ts">
import { useData, withBase } from 'vitepress'
import DefaultTheme from 'vitepress/theme'

const { page } = useData()
</script>

<template>
  <DefaultTheme.Layout>
    <template #doc-before>
      <div v-if="page.filePath === 'README.md' || page.relativePath === 'index.md'" class="guide-lead">
        <div class="guide-lead-main">
          <p class="guide-path">注册事件 <span aria-hidden="true">→</span> 读取消息 <span aria-hidden="true">→</span> 回复会话</p>
          <a class="guide-start" :href="withBase('/getting-started/quick-start.html')">运行第一个机器人 <span aria-hidden="true">↗</span></a>
        </div>
        <code>ctx.ReplyText("你好")</code>
      </div>
    </template>
  </DefaultTheme.Layout>
</template>
