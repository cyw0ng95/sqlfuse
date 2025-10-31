<template>
  <v-container class="pa-6" fluid>
    <v-row>
      <v-col cols="12">
        <v-card class="mb-6 modern-card" elevation="8">
          <v-card-title class="d-flex align-center bg-gradient-primary">
            <v-icon class="mr-2">mdi-heart-pulse</v-icon>
            Health Status
          </v-card-title>
          <v-card-text class="pa-4">
            <div v-if="health.error" class="text-error">
              <v-icon class="mr-2" color="error">mdi-alert-circle</v-icon>
              {{ health.error }}
            </div>
            <div v-else>
              <v-chip class="mb-2" color="success" prepend-icon="mdi-check-circle">
                {{ health.status || 'unknown' }}
              </v-chip>
              <div v-if="health.timestamp" class="text-caption text-medium-emphasis mt-2">
                <v-icon class="mr-1" size="small">mdi-clock-outline</v-icon>
                {{ health.timestamp }}
              </div>
            </div>
          </v-card-text>
          <v-card-actions class="pa-4 pt-0">
            <v-btn color="primary" prepend-icon="mdi-refresh" variant="tonal" @click="getHealth">
              Refresh
            </v-btn>
          </v-card-actions>
        </v-card>

        <v-card class="mb-6 modern-card" elevation="8">
          <v-card-title class="d-flex align-center bg-gradient-info">
            <v-icon class="mr-2">mdi-information</v-icon>
            System Information
          </v-card-title>
          <v-card-text class="pa-4">
            <div v-if="info.error" class="text-error">
              <v-icon class="mr-2" color="error">mdi-alert-circle</v-icon>
              {{ info.error }}
            </div>
            <div v-else>
              <pre class="info-pre">{{ formattedInfo }}</pre>
            </div>
          </v-card-text>
          <v-card-actions class="pa-4 pt-0">
            <v-btn color="info" prepend-icon="mdi-refresh" variant="tonal" @click="getInfo">
              Refresh
            </v-btn>
          </v-card-actions>
        </v-card>

        <!-- Job manager component -->
        <JobManager />

        <v-card class="modern-card" elevation="8">
          <v-card-title class="d-flex align-center bg-gradient-secondary">
            <v-icon class="mr-2">mdi-code-braces</v-icon>
            SQL Generators
          </v-card-title>
          <v-card-text class="pa-4">
            <div v-if="generators.error" class="text-error">
              <v-icon class="mr-2" color="error">mdi-alert-circle</v-icon>
              {{ generators.error }}
            </div>
            <div v-else>
              <div class="mb-3">
                <v-chip color="primary" prepend-icon="mdi-cog" variant="outlined">
                  {{ generators.generator || 'n/a' }}
                </v-chip>
              </div>
              <div v-if="generators.stmts">
                <div class="text-subtitle-2 mb-2 font-weight-bold">Supported Statements:</div>
                <pre class="info-pre">{{ formattedGenerators }}</pre>
              </div>
            </div>
          </v-card-text>
          <v-card-actions class="pa-4 pt-0">
            <v-btn color="secondary" prepend-icon="mdi-refresh" variant="tonal" @click="getGenerators">
              Refresh
            </v-btn>
          </v-card-actions>
        </v-card>

      </v-col>
    </v-row>
  </v-container>
</template>

<script setup>
  import { computed, onMounted, ref } from 'vue'
  import JobManager from '@/components/JobManager.vue'

  const health = ref({})
  const info = ref({})
  const generators = ref({})

  async function fetchJson (path) {
    const API_BASE = 'http://localhost:8080'
    const url = path.startsWith('http://') || path.startsWith('https://') ? path : `${API_BASE}${path}`
    try {
      const res = await fetch(url, { cache: 'no-store' })
      if (!res.ok) {
        return { error: `HTTP ${res.status} ${res.statusText}` }
      }
      return await res.json()
    } catch (error) {
      return { error: error.message || String(error) }
    }
  }

  async function getHealth () {
    const data = await fetchJson('/health')
    health.value = data
  }

  async function getInfo () {
    const data = await fetchJson('/info')
    info.value = data
  }

  async function getGenerators () {
    const data = await fetchJson('/generators/get')
    generators.value = data
  }

  const formattedInfo = computed(() => {
    try {
      return JSON.stringify(info.value, null, 2)
    } catch {
      return String(info.value)
    }
  })

  // Render the generators.stmts map (stmt -> weight) into a readable string
  const formattedGenerators = computed(() => {
    const stmts = generators.value && generators.value.stmts ? generators.value.stmts : null
    if (!stmts) return ''

    // Convert to entries and ensure numeric weights
    const entries = Object.entries(stmts).map(([k, v]) => [k, Number(v || 0)])
    const total = entries.reduce((s, [, w]) => s + w, 0)

    // Sort by weight desc, then name
    entries.sort((a, b) => {
      const wdiff = b[1] - a[1]
      if (wdiff !== 0) return wdiff
      return a[0].localeCompare(b[0])
    })

    const lines = entries.map(([k, w]) => {
      const pct = total > 0 ? ((w / total) * 100).toFixed(1) : '0.0'
      return `${k}: ${w} (${pct}%)`
    })

    lines.unshift(`Total tokens: ${total}`)
    return lines.join('\n')
  })

  onMounted(() => {
    getHealth()
    getInfo()
    getGenerators()
  })
</script>

<style scoped>
.modern-card {
  border-radius: 12px !important;
  overflow: hidden;
  transition: transform 0.2s ease-in-out;
}

.modern-card:hover {
  transform: translateY(-2px);
}

.bg-gradient-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white !important;
}

.bg-gradient-info {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white !important;
}

.bg-gradient-secondary {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: white !important;
}

.info-pre {
  background-color: #f5f5f5;
  border-radius: 8px;
  padding: 16px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  border: 1px solid #e0e0e0;
}

.text-error {
  color: #d32f2f;
  display: flex;
  align-items: center;
}
</style>
