<template>
  <v-card class="mb-6 modern-card" elevation="8">
    <v-card-title class="d-flex align-center bg-gradient-jobs">
      <v-icon class="mr-2">mdi-briefcase</v-icon>
      Job Management
    </v-card-title>
    <v-card-text class="pa-4">
      <v-row>
        <v-col cols="12" md="8">
          <v-select
            v-model="selectedExecutor"
            density="comfortable"
            :disabled="executors.length === 0"
            hide-details
            item-title="executor"
            item-value="executor"
            :items="executors"
            label="Executor"
            prepend-inner-icon="mdi-cog"
            variant="outlined"
          >
            <template #item="{ props, item }">
              <v-list-item v-bind="props">
                <template #prepend>
                  <v-icon>mdi-engine</v-icon>
                </template>
                <template #title>
                  {{ item.raw.executor }}
                </template>
                <template v-if="item.raw.flavor" #subtitle>
                  Flavor: {{ item.raw.flavor }}
                </template>
              </v-list-item>
            </template>
            <template #selection="{ item }">
              <span>{{ item.raw.executor }}</span>
              <span v-if="item.raw.flavor" class="text-caption ml-2 text-primary">({{ item.raw.flavor }})</span>
            </template>
          </v-select>
          <v-text-field
            v-model="args"
            class="mt-3"
            density="comfortable"
            hide-details
            label="Arguments (separated by space)"
            placeholder="--workers 4 --queries 100"
            prepend-inner-icon="mdi-code-tags"
            variant="outlined"
          />
        </v-col>
        <v-col class="d-flex align-center" cols="12" md="4">
          <v-btn
            block
            color="primary"
            elevation="2"
            :loading="creating"
            prepend-icon="mdi-play-circle"
            size="large"
            @click="createJob"
          >
            Start Job
          </v-btn>
        </v-col>
      </v-row>

      <div v-if="createError" class="error-alert mt-3">
        <v-icon class="mr-2" color="error">mdi-alert-circle</v-icon>
        {{ createError }}
      </div>
      <div v-if="lastID" class="success-alert mt-3">
        <v-icon class="mr-2" color="success">mdi-check-circle</v-icon>
        Started job ID: <strong>{{ lastID }}</strong>
      </div>

      <v-divider class="my-6" />

      <div class="text-subtitle-1 font-weight-bold mb-3">
        <v-icon class="mr-2">mdi-magnify</v-icon>
        Query Job Status
      </div>

      <v-row>
        <v-col cols="12" md="6">
          <v-text-field
            v-model="queryID"
            density="comfortable"
            hide-details
            label="Job ID"
            prepend-inner-icon="mdi-identifier"
            variant="outlined"
          />
        </v-col>
        <v-col class="d-flex align-center gap-2" cols="12" md="6">
          <v-btn color="info" prepend-icon="mdi-information" variant="tonal" @click="fetchStatus">
            Status
          </v-btn>
          <v-btn color="info" prepend-icon="mdi-text-box" variant="tonal" @click="fetchInfo">
            Info
          </v-btn>
          <v-btn 
            v-if="!streamingLogs && queryID" 
            color="success" 
            prepend-icon="mdi-broadcast" 
            variant="tonal" 
            @click="startStreaming"
          >
            Stream Logs
          </v-btn>
          <v-btn 
            v-if="streamingLogs" 
            color="warning" 
            prepend-icon="mdi-broadcast-off" 
            variant="tonal" 
            @click="disconnectLogStream"
          >
            Stop Stream
          </v-btn>
          <v-btn color="error" prepend-icon="mdi-stop-circle" variant="tonal" @click="stopJob">
            Stop
          </v-btn>
        </v-col>
      </v-row>

      <div v-if="status" class="status-box mt-4">
        <v-chip class="mb-2" color="info" prepend-icon="mdi-information">
          {{ status.status }}
          <span v-if="status.pid" class="ml-2">(PID: {{ status.pid }})</span>
        </v-chip>
        <div v-if="status.started_at" class="text-caption">
          <v-icon size="small">mdi-clock-start</v-icon> Started: {{ status.started_at }}
        </div>
        <div v-if="status.ended_at" class="text-caption">
          <v-icon size="small">mdi-clock-end</v-icon> Ended: {{ status.ended_at }}
        </div>
        <div v-if="status.exit_code != null" class="text-caption">
          <v-icon size="small">mdi-exit-to-app</v-icon> Exit Code: {{ status.exit_code }}
        </div>
        <div v-if="status.signal" class="text-caption error-text">
          <v-icon size="small" color="error">mdi-alert-octagon</v-icon> Killed by Signal: {{ status.signal }}
        </div>
      </div>

      <!-- Streaming logs display -->
      <div v-if="streamingLogs || streamLogs.length > 0" class="mt-4">
        <v-card elevation="2">
          <v-card-title class="d-flex align-center bg-gradient-stream">
            <v-icon class="mr-2">mdi-broadcast</v-icon>
            Streaming Logs
            <v-spacer />
            <v-chip v-if="streamingLogs" color="success" size="small" variant="flat">
              <v-icon start>mdi-circle</v-icon>
              Live
            </v-chip>
          </v-card-title>
          <v-card-text class="pa-0">
            <div class="stream-container">
              <pre class="stream-pre" v-for="(log, index) in streamLogs" :key="index"><span :class="`log-${log.stream}`">{{ log.data }}</span></pre>
            </div>
          </v-card-text>
        </v-card>
      </div>

      <div v-if="infoData" class="mt-4">
        <v-expansion-panels>
          <v-expansion-panel>
            <v-expansion-panel-title>
              <v-icon class="mr-2">mdi-console</v-icon>
              Standard Output
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <pre class="output-pre">{{ infoData.stdout || '(empty)' }}</pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
          <v-expansion-panel>
            <v-expansion-panel-title>
              <v-icon class="mr-2">mdi-alert</v-icon>
              Standard Error
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <pre class="output-pre error-output">{{ infoData.stderr || '(empty)' }}</pre>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>
      </div>

    </v-card-text>
    <v-card-actions class="pa-4 pt-0">
      <v-btn color="secondary" prepend-icon="mdi-broom" variant="text" @click="clear">
        Clear
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script setup>
  import { onMounted, onUnmounted, ref } from 'vue'
  import { API_BASE_URL, WS_BASE_URL } from '../config.js'

  const props = defineProps({
    customWeights: {
      type: Object,
      default: null
    }
  })

  const executors = ref([])
  const selectedExecutor = ref('')
  const args = ref('')
  const creating = ref(false)
  const createError = ref('')
  const lastID = ref('')

  const queryID = ref('')
  const status = ref(null)
  const infoData = ref(null)
  
  // WebSocket streaming support
  const streamingLogs = ref(false)
  const streamLogs = ref([])
  let logWebSocket = null

  async function fetchJson (path, opts) {
    const url = path.startsWith('http://') || path.startsWith('https://') ? path : `${API_BASE_URL}${path}`
    try {
      const res = await fetch(url, opts)
      if (!res.ok) {
        const text = await res.text()
        return { error: `HTTP ${res.status} ${res.statusText}: ${text}` }
      }
      return await res.json()
    } catch (error) {
      return { error: error.message || String(error) }
    }
  }

  async function getExecutors () {
    const data = await fetchJson('/executors')
    if (data && data.error) {
      executors.value = []
      return
    }
    // data expected to be an array of { executor, path }
    executors.value = Array.isArray(data) ? data : []
  }

  onMounted(() => {
    getExecutors()
  })

  onUnmounted(() => {
    disconnectLogStream()
  })

  function connectLogStream(jobID) {
    // Disconnect existing stream if any
    disconnectLogStream()
    
    streamLogs.value = []
    streamingLogs.value = true
    
    const wsUrl = `${WS_BASE_URL}/job/logs/stream?id=${encodeURIComponent(jobID)}`
    logWebSocket = new WebSocket(wsUrl)
    
    logWebSocket.onopen = () => {
      console.log('WebSocket connected for job', jobID)
    }
    
    logWebSocket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        streamLogs.value.push({
          stream: msg.stream,
          data: msg.data,
          timestamp: new Date().toISOString()
        })
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }
    
    logWebSocket.onerror = (error) => {
      console.error('WebSocket error:', error)
      streamingLogs.value = false
    }
    
    logWebSocket.onclose = () => {
      console.log('WebSocket closed')
      streamingLogs.value = false
      logWebSocket = null
    }
  }
  
  function disconnectLogStream() {
    if (logWebSocket) {
      logWebSocket.close()
      logWebSocket = null
    }
    streamingLogs.value = false
  }

  async function createJob () {
    createError.value = ''
    if (!selectedExecutor.value) {
      createError.value = 'executor is required'
      return
    }
    creating.value = true
    // split args by whitespace
    const argsList = args.value.trim() === '' ? [] : (args.value.trim().match(/\S+/g) || [])
    const payload = { executor: selectedExecutor.value, args: argsList }
    
    // Add custom weights if provided
    if (props.customWeights && Object.keys(props.customWeights).length > 0) {
      payload.weights = props.customWeights
    }
    
    const data = await fetchJson('/job/new', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) })
    creating.value = false
    if (data.error) {
      createError.value = data.error
      return
    }
    lastID.value = data.id || ''
    queryID.value = lastID.value
    
    // Start streaming logs for the new job
    if (lastID.value) {
      connectLogStream(lastID.value)
    }
  }

  async function fetchStatus () {
    if (!queryID.value) return
    const data = await fetchJson(`/job/status?id=${encodeURIComponent(queryID.value)}`)
    if (data.error) {
      status.value = { error: data.error }
      return
    }
    status.value = data
  }

  async function fetchInfo () {
    if (!queryID.value) return
    const data = await fetchJson(`/job/info?id=${encodeURIComponent(queryID.value)}`)
    if (data.error) {
      infoData.value = { error: data.error }
      return
    }
    infoData.value = data
  }

  async function stopJob () {
    if (!queryID.value) return
    const data = await fetchJson(`/job/stop?id=${encodeURIComponent(queryID.value)}`, { method: 'POST' })
    if (data.error) {
      status.value = { error: data.error }
      return
    }
    // refresh status
    await fetchStatus()
    
    // Disconnect stream when job is stopped
    disconnectLogStream()
  }
  
  function startStreaming () {
    if (!queryID.value) return
    connectLogStream(queryID.value)
  }

  function clear () {
    selectedExecutor.value = ''
    args.value = ''
    lastID.value = ''
    queryID.value = ''
    status.value = null
    infoData.value = null
    createError.value = ''
    streamLogs.value = []
    disconnectLogStream()
  }
</script>

<style scoped>
.modern-card {
  border-radius: 12px !important;
  overflow: hidden;
}

.bg-gradient-jobs {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: white !important;
}

.error-alert {
  color: #d32f2f;
  background-color: #ffebee;
  padding: 12px 16px;
  border-radius: 8px;
  border-left: 4px solid #d32f2f;
  display: flex;
  align-items: center;
}

.success-alert {
  color: #2e7d32;
  background-color: #e8f5e9;
  padding: 12px 16px;
  border-radius: 8px;
  border-left: 4px solid #4caf50;
  display: flex;
  align-items: center;
}

.status-box {
  background-color: #f5f5f5;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.error-text {
  color: #d32f2f;
  font-weight: 500;
}

.output-pre {
  background-color: #263238;
  color: #aed581;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  margin: 0;
}

.error-output {
  color: #ef9a9a;
}

.gap-2 {
  gap: 8px;
}

.bg-gradient-stream {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white !important;
}

.stream-container {
  max-height: 500px;
  overflow-y: auto;
  background-color: #1e1e1e;
  padding: 16px;
}

.stream-pre {
  color: #d4d4d4;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.log-stdout {
  color: #4ec9b0;
}

.log-stderr {
  color: #f48771;
}

.log-status {
  color: #dcdcaa;
  font-weight: bold;
}
</style>
