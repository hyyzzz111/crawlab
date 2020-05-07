<template>
  <div class="app-container">
    <!--overall metrics-->
    <el-row>
      <ul class="metric-list">
        <li class="metric-item" v-for="m in metrics" @click="onClickMetric(m)" :key="m.name">
          <div class="metric-icon" :class="m.color">
            <!--            <font-awesome-icon :icon="m.icon"/>-->
            <i :class="m.icon"></i>
          </div>
          <div class="metric-content" :class="m.color">
            <div class="metric-wrapper">
              <div class="metric-number">
                {{overviewStats[m.name]}}
              </div>
              <div class="metric-name">
                {{$t(m.label)}}
              </div>
            </div>
          </div>
        </li>
      </ul>
    </el-row>
    <!--./overall metrics-->

    <!--performance metrics-->
    <el-row>
      <ul class="performance-metric-list">
        <li class="performance-metric-item mongo">
          <div class="performance-metric-title">
            <i class="fa fa-database"></i>
            MongoDB
          </div>
          <div class="performance-metric-body">
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Disk')}}
              </span>
              <el-progress
                :stroke-width="20"
                :percentage="mongoDiskPercent"
                text-inside
                :status="getProgressStatus(mongoDiskPercent)"
              />
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Memory')}}
              </span>
              <el-progress
                :stroke-width="20"
                :percentage="mongoMemoryPercent"
                text-inside
                :status="getProgressStatus(mongoMemoryPercent)"
              />
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Storage Size')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{mongoStorageSize}} GB
              </el-tag>
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Index Size')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{mongoIndexSize}} GB
              </el-tag>
            </div>
          </div>
        </li>
        <li class="performance-metric-item redis">
          <div class="performance-metric-title">
            <i class="fa fa-database"></i>
            Redis
          </div>
          <div class="performance-metric-body">
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Total Allocated')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{redisTotalAllocated}} MB
              </el-tag>
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Peak Allocated')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{redisPeakAllocated}} MB
              </el-tag>
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Dataset Size')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{redisDataset}} MB
              </el-tag>
            </div>
            <div class="progress-item">
              <span class="progress-label">
                {{$t('Overhead Size')}}
              </span>
              <el-tag
                text-inside
                size="small"
                type="primary"
              >
                {{redisOverhead}} MB
              </el-tag>
            </div>
          </div>
        </li>
      </ul>
    </el-row>
    <!--./overall metrics-->

    <!--performance metrics-->
    <el-row>
      <el-card shadow="hover">
        <h4 class="title">{{$t('Daily New Tasks')}}</h4>
        <div id="echarts-daily-tasks" class="echarts-box"></div>
      </el-card>
    </el-row>
  </div>
</template>

<script>
import echarts from 'echarts'

export default {
  name: 'Home',
  data () {
    return {
      // echarts instance
      echarts: {},
      // overall stats
      overviewStats: {},
      dailyTasks: [],
      metrics: [
        { name: 'task_count', label: 'Total Tasks', icon: 'fa fa-check', color: 'blue', path: 'tasks' },
        { name: 'spider_count', label: 'Spiders', icon: 'fa fa-bug', color: 'green', path: 'spiders' },
        { name: 'active_node_count', label: 'Active Nodes', icon: 'fa fa-server', color: 'red', path: 'nodes' },
        { name: 'schedule_count', label: 'Schedules', icon: 'fa fa-clock-o', color: 'orange', path: 'schedules' },
        { name: 'project_count', label: 'Projects', icon: 'fa fa-code-fork', color: 'grey', path: 'projects' }
      ],
      // mongo related
      mongoStats: {
        db_stats: {},
        mem_stats: {}
      },
      redisStats: {}
    }
  },
  computed: {
    mongoDiskPercent () {
      return Math.round(this.mongoStats.db_stats.fsUsedSize / this.mongoStats.db_stats.fsTotalSize * 100)
    },
    mongoMemoryPercent () {
      return Math.round(this.mongoStats.mem_stats.resident / this.mongoStats.mem_stats.virtual * 100)
    },
    mongoStorageSize () {
      const value = this.mongoStats.db_stats.storageSize / 1024 / 1024 / 1024 * 0.01
      if (value < 0.01) return '<0.01'
      return value.toFixed(2)
    },
    mongoIndexSize () {
      const value = this.mongoStats.db_stats.indexSize / 1024 / 1024 / 1024
      if (value < 0.01) return '<0.01'
      return value.toFixed(2)
    },
    redisTotalAllocated () {
      return (this.redisStats.total_allocated / 1024 / 1024).toFixed(2)
    },
    redisPeakAllocated () {
      return (this.redisStats.peak_allocated / 1024 / 1024).toFixed(2)
    },
    redisDataset () {
      return (this.redisStats.dataset_bytes / 1024 / 1024).toFixed(2)
    },
    redisOverhead () {
      return (this.redisStats.overhead_total / 1024 / 1024).toFixed(2)
    }
  },
  methods: {
    initEchartsDailyTasks () {
      const option = {
        xAxis: {
          type: 'category',
          data: this.dailyTasks.map(d => d.date)
        },
        yAxis: {
          type: 'value'
        },
        series: [{
          data: this.dailyTasks.map(d => d.task_count),
          type: 'line',
          areaStyle: {},
          smooth: true
        }],
        tooltip: {
          trigger: 'axis',
          show: true
        }
      }
      this.echarts.dailyTasks = echarts.init(this.$el.querySelector('#echarts-daily-tasks'))
      this.echarts.dailyTasks.setOption(option)
    },
    onClickMetric (m) {
      this.$router.push(`/${m.path}`)
    },
    async getBasicStats () {
      const res = await this.$request.get('/stats/home')

      // overview stats
      this.overviewStats = res.data.data.overview

      // daily tasks
      this.dailyTasks = res.data.data.daily
      this.initEchartsDailyTasks()
    },
    async getMonitorStats () {
      await this.getMongoStats()
      await this.getRedisStats()
      await this.getNodesStats()
    },
    async getMongoStats () {
      const res = await this.$request.get('/monitor/mongo')
      this.mongoStats = res.data.data
    },
    async getRedisStats () {
      const res = await this.$request.get('/monitor/redis')
      this.redisStats = res.data.data
    },
    async getNodesStats () {
      const res = await this.$request.get('/nodes')
      res.data.data.forEach(async d => {
        const res = await this.$request.get('/monitor/nodes/' + d._id)
        console.log(res)
      })
    },
    getProgressStatus (value) {
      if (value >= 80) {
        return 'exception'
      } else if (value >= 40) {
        return 'warning'
      } else {
        return 'success'
      }
    }
  },
  async created () {
    await this.getBasicStats()
    await this.getMonitorStats()
  },
  mounted () {
  }
}
</script>

<style scoped lang="scss">
  .metric-list {
    margin-top: 0;
    padding-left: 0;
    list-style: none;
    display: flex;
    font-size: 16px;

    .metric-item:last-child .metric-card {
      margin-right: 0;
    }

    .metric-item:hover {
      transform: scale(1.05);
      transition: transform 0.5s ease;
    }

    .metric-item {
      flex-basis: 20%;
      height: 64px;
      display: flex;
      color: white;
      cursor: pointer;
      transform: scale(1);
      transition: transform 0.5s ease;

      .metric-icon {
        display: inline-flex;
        width: 64px;
        align-items: center;
        justify-content: center;
        border-top-left-radius: 5px;
        border-bottom-left-radius: 5px;
        font-size: 24px;

        svg {
          width: 24px;
        }
      }

      .metric-content {
        display: flex;
        width: calc(100% - 80px);
        align-items: center;
        opacity: 0.85;
        font-size: 14px;
        padding-left: 15px;
        border-top-right-radius: 5px;
        border-bottom-right-radius: 5px;

        .metric-number {
          font-weight: bolder;
          margin-bottom: 5px;
        }
      }

      .metric-icon.blue,
      .metric-content.blue {
        background: #409eff;
      }

      .metric-icon.green,
      .metric-content.green {
        background: #67c23a;
      }

      .metric-icon.red,
      .metric-content.red {
        background: #f56c6c;
      }

      .metric-icon.orange,
      .metric-content.orange {
        background: #E6A23C;
      }

      .metric-icon.grey,
      .metric-content.grey {
        background: #97a8be;
      }
    }
  }

  .performance-metric-list {
    list-style: none;
    display: flex;
    margin: 0 0 20px;
    padding: 0;

    .performance-metric-item {
      width: 270px;
      height: 270px;
      border: 1px solid #EBEEF5;
      border-radius: 5px;
      margin-right: 20px;

      .performance-metric-title {
        border-top-left-radius: 5px;
        border-top-right-radius: 5px;
        display: flex;
        align-items: center;
        justify-content: center;
        height: 60px;
        color: white;

        i.fa {
          margin-right: 5px;
        }
      }

      .performance-metric-body {
        padding: 20px;

        .progress-item {
          display: flex;
          align-items: center;
          margin-bottom: 20px;

          .progress-label {
            flex-basis: 80px;
            display: inline-block;
            text-align: right;
            padding-right: 10px;
            font-size: 14px;
            color: #5a5e66;
          }

          .el-progress {
            flex-basis: calc(100% - 80px);
            display: inline-block;
          }
        }
      }
    }

    .performance-metric-item.mongo .performance-metric-title {
      background: #67c23a;
    }

    .performance-metric-item.redis .performance-metric-title {
      background: #f56c6c;
    }
  }

  .title {
    padding: 0;
    margin: 0;
  }

  #echarts-daily-tasks {
    height: 360px;
    width: 100%;
  }

  .el-card {
    /*border: 1px solid lightgrey;*/
  }

  .svg-inline--fa {
    width: 100%;
    height: 100%;
  }
</style>
