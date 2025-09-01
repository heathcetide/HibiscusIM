// 监控系统JavaScript文件
class MonitorSystem {
  constructor() {
    this.data = {
      activeTab: 'overview',
      monitorData: {},
      slowQueries: [],
      traces: [],
      systemStats: null,
      loading: false,
      refreshInterval: null
    }
    
    this.init()
  }

  init() {
    this.loadMonitorData()
    this.startAutoRefresh()
    this.bindEvents()
  }

  bindEvents() {
    // 绑定标签页切换事件
    document.addEventListener('click', (e) => {
      if (e.target.matches('[data-tab]')) {
        const tab = e.target.dataset.tab
        this.switchTab(tab)
      }
    })

    // 绑定刷新按钮事件
    const refreshBtn = document.querySelector('[data-action="refresh"]')
    if (refreshBtn) {
      refreshBtn.addEventListener('click', () => this.refreshData())
    }
  }

  async loadMonitorData() {
    try {
      this.data.loading = true
      
      // 加载系统概览
      const overviewResp = await fetch('/api/admin/monitor/overview')
      const overviewData = await overviewResp.json()
      if (overviewData.success) {
        this.data.monitorData = overviewData.data
      }

      // 加载慢查询
      const slowResp = await fetch('/api/admin/monitor/sql/slow?limit=20')
      const slowData = await slowResp.json()
      if (slowData.success) {
        this.data.slowQueries = slowData.data || []
      }

      // 加载链路追踪
      const tracesResp = await fetch('/api/admin/monitor/traces?limit=20')
      const tracesData = await tracesResp.json()
      if (tracesData.success) {
        this.data.traces = tracesData.data || []
      }

      // 加载系统统计
      const systemResp = await fetch('/api/admin/monitor/system/latest')
      const systemData = await systemResp.json()
      if (systemData.success) {
        this.data.systemStats = systemData.data
      }

      this.updateUI()
      
    } catch (error) {
      console.error('加载监控数据失败:', error)
      this.showError('加载监控数据失败')
    } finally {
      this.data.loading = false
    }
  }

  switchTab(tabName) {
    this.data.activeTab = tabName
    
    // 隐藏所有标签页内容
    document.querySelectorAll('[data-tab-content]').forEach(content => {
      content.style.display = 'none'
    })
    
    // 显示选中的标签页内容
    const activeContent = document.querySelector(`[data-tab-content="${tabName}"]`)
    if (activeContent) {
      activeContent.style.display = 'block'
    }
    
    // 更新标签页样式
    document.querySelectorAll('[data-tab]').forEach(tab => {
      tab.classList.remove('border-indigo-500', 'text-indigo-600')
      tab.classList.add('border-transparent', 'text-gray-500')
    })
    
    const activeTab = document.querySelector(`[data-tab="${tabName}"]`)
    if (activeTab) {
      activeTab.classList.remove('border-transparent', 'text-gray-500')
      activeTab.classList.add('border-indigo-500', 'text-indigo-600')
    }
  }

  updateUI() {
    // 更新概览卡片
    this.updateOverviewCards()
    
    // 更新标签页内容
    this.updateTabContent()
  }

  updateOverviewCards() {
    // 更新系统状态
    const statusCard = document.querySelector('[data-card="status"]')
    if (statusCard && this.data.monitorData.overview?.system?.status) {
      statusCard.textContent = this.data.monitorData.overview.system.status
    }

    // 更新CPU使用率
    const cpuCard = document.querySelector('[data-card="cpu"]')
    if (cpuCard && this.data.systemStats?.cpu?.usage_percent) {
      cpuCard.textContent = this.data.systemStats.cpu.usage_percent.toFixed(1) + '%'
    }

    // 更新内存使用率
    const memoryCard = document.querySelector('[data-card="memory"]')
    if (memoryCard && this.data.systemStats?.memory?.usage_percent) {
      memoryCard.textContent = this.data.systemStats.memory.usage_percent.toFixed(1) + '%'
    }

    // 更新慢查询数量
    const slowQueryCard = document.querySelector('[data-card="slow-queries"]')
    if (slowQueryCard) {
      slowQueryCard.textContent = this.data.slowQueries.length
    }
  }

  updateTabContent() {
    switch (this.data.activeTab) {
      case 'overview':
        this.updateOverviewTab()
        break
      case 'sql':
        this.updateSQLTab()
        break
      case 'traces':
        this.updateTracesTab()
        break
      case 'system':
        this.updateSystemTab()
        break
    }
  }

  updateOverviewTab() {
    const overviewContent = document.querySelector('[data-tab-content="overview"]')
    if (!overviewContent) return

    const html = `
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <h4 class="font-medium text-gray-700 mb-2">基本信息</h4>
          <dl class="space-y-2">
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">监控状态</dt>
              <dd class="text-sm text-gray-900">${this.data.monitorData.enabled ? '已启用' : '已禁用'}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">链路追踪</dt>
              <dd class="text-sm text-gray-900">${this.data.traces.length} 条</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">慢查询</dt>
              <dd class="text-sm text-gray-900">${this.data.slowQueries.length} 条</dd>
            </div>
          </dl>
        </div>
        <div>
          <h4 class="font-medium text-gray-700 mb-2">配置信息</h4>
          <dl class="space-y-2">
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">慢查询阈值</dt>
              <dd class="text-sm text-gray-900">${this.data.monitorData.config?.slow_threshold || '100ms'}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">最大跨度数</dt>
              <dd class="text-sm text-gray-900">${this.data.monitorData.config?.max_spans || '10000'}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">监控间隔</dt>
              <dd class="text-sm text-gray-900">${this.data.monitorData.config?.monitor_interval || '30s'}</dd>
            </div>
          </dl>
        </div>
      </div>
    `
    
    overviewContent.innerHTML = html
  }

  updateSQLTab() {
    const sqlContent = document.querySelector('[data-tab-content="sql"]')
    if (!sqlContent) return

    if (this.data.slowQueries.length === 0) {
      sqlContent.innerHTML = '<p class="text-gray-500">暂无慢查询数据</p>'
      return
    }

    const html = `
      <div class="space-y-4">
        <div>
          <h4 class="font-medium text-gray-700 mb-2">慢查询列表</h4>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">SQL</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">表</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">耗时</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">时间</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                ${this.data.slowQueries.map(query => `
                  <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${query.sql}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${query.table}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-red-600 font-medium">${query.duration}ms</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${new Date(query.timestamp).toLocaleString()}</td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    `
    
    sqlContent.innerHTML = html
  }

  updateTracesTab() {
    const tracesContent = document.querySelector('[data-tab-content="traces"]')
    if (!tracesContent) return

    if (this.data.traces.length === 0) {
      tracesContent.innerHTML = '<p class="text-gray-500">暂无链路追踪数据</p>'
      return
    }

    const html = `
      <div class="space-y-4">
        <div>
          <h4 class="font-medium text-gray-700 mb-2">追踪列表</h4>
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200">
              <thead class="bg-gray-50">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">追踪ID</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">操作名</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">耗时</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
                </tr>
              </thead>
              <tbody class="bg-white divide-y divide-gray-200">
                ${this.data.traces.map(span => `
                  <tr>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${span.traceID}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${span.name}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${span.duration}ms</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                      <span class="inline-flex px-2 py-1 text-xs font-semibold rounded-full ${span.error ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'}">
                        ${span.error ? '错误' : '成功'}
                      </span>
                    </td>
                  </tr>
                `).join('')}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    `
    
    tracesContent.innerHTML = html
  }

  updateSystemTab() {
    const systemContent = document.querySelector('[data-tab-content="system"]')
    if (!systemContent) return

    if (!this.data.systemStats) {
      systemContent.innerHTML = '<p class="text-gray-500">暂无系统监控数据</p>'
      return
    }

    const html = `
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <h4 class="font-medium text-gray-700 mb-2">CPU信息</h4>
          <dl class="space-y-2">
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">使用率</dt>
              <dd class="text-sm text-gray-900">${this.data.systemStats.cpu?.usage_percent?.toFixed(1) || '0'}%</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">核心数</dt>
              <dd class="text-sm text-gray-900">${this.data.systemStats.cpu?.cores || '0'}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">频率</dt>
              <dd class="text-sm text-gray-900">${this.data.systemStats.cpu?.frequency?.toFixed(0) || '0'} MHz</dd>
            </div>
          </dl>
        </div>
        <div>
          <h4 class="font-medium text-gray-700 mb-2">内存信息</h4>
          <dl class="space-y-2">
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">总内存</dt>
              <dd class="text-sm text-gray-900">${this.formatBytes(this.data.systemStats.memory?.total)}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">已使用</dt>
              <dd class="text-sm text-gray-900">${this.formatBytes(this.data.systemStats.memory?.used)}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-sm text-gray-500">可用</dt>
              <dd class="text-sm text-gray-900">${this.formatBytes(this.data.systemStats.memory?.available)}</dd>
            </div>
          </dl>
        </div>
      </div>
    `
    
    systemContent.innerHTML = html
  }

  refreshData() {
    this.loadMonitorData()
    this.showSuccess('监控数据已刷新')
  }

  startAutoRefresh() {
    // 每30秒自动刷新一次数据
    this.data.refreshInterval = setInterval(() => {
      this.loadMonitorData()
    }, 30000)
  }

  stopAutoRefresh() {
    if (this.data.refreshInterval) {
      clearInterval(this.data.refreshInterval)
      this.data.refreshInterval = null
    }
  }

  formatBytes(bytes) {
    if (!bytes) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  showSuccess(message) {
    // 显示成功消息
    if (window.Alpine && Alpine.store('toasts')) {
      Alpine.store('toasts').info(message)
    } else {
      console.log('Success:', message)
    }
  }

  showError(message) {
    // 显示错误消息
    if (window.Alpine && Alpine.store('toasts')) {
      Alpine.store('toasts').error(message)
    } else {
      console.error('Error:', message)
    }
  }

  destroy() {
    this.stopAutoRefresh()
  }
}

// 初始化监控系统
document.addEventListener('DOMContentLoaded', () => {
  window.monitorSystem = new MonitorSystem()
})

// 导出类
window.MonitorSystem = MonitorSystem
