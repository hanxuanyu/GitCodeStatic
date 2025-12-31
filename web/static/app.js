const { createApp } = Vue;
const { ElMessage, ElMessageBox } = ElementPlus;

const API_BASE = '/api/v1';

createApp({
    data() {
        return {
            activeTab: 'repos',
            repos: [],
            reposLoading: false,
            statsLoading: false,
            addRepoVisible: false,
            switchBranchVisible: false,
            addRepoForm: {
                urls: '',
                branch: 'main',
                username: '',
                password: ''
            },
            switchBranchForm: {
                branch: '',
                repoId: null,
                repoUrl: '',
                currentBranch: ''
            },
            branches: [],
            branchesLoading: false,
            statsForm: {
                repo_id: null,
                branch: 'main',
                constraint_type: 'commit_limit',
                from: '',
                to: '',
                limit: 100
            },
            statsDateRange: null,
            statsFormBranches: [],
            statsFormBranchesLoading: false,
            selectedStatsResult: null,
            tasks: [],
            tasksLoading: false,
            caches: [],
            cachesLoading: false
        };
    },
    mounted() {
        this.loadRepos();
        this.loadCaches();
    },
    watch: {
        activeTab(newTab) {
            if (newTab === 'tasks') {
                this.loadTasks();
            } else if (newTab === 'caches') {
                this.loadCaches();
            }
        }
    },
    methods: {
        async loadRepos() {
            this.reposLoading = true;
            try {
                const response = await axios.get(`${API_BASE}/repos`);
                if (response.data.code === 0) {
                    this.repos = response.data.data.repositories || [];
                } else {
                    ElMessage.error(response.data.message || '加载仓库列表失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            } finally {
                this.reposLoading = false;
            }
        },
        showAddRepoDialog() {
            this.addRepoForm = { 
                urls: '', 
                branch: 'main',
                username: '',
                password: ''
            };
            this.addRepoVisible = true;
        },
        async onRepoChange() {
            this.statsForm.branch = 'main';
            this.statsFormBranches = [];
            if (this.statsForm.repo_id) {
                await this.loadStatsFormBranches();
            }
        },
        async loadStatsFormBranches() {
            if (!this.statsForm.repo_id) return;
            
            this.statsFormBranchesLoading = true;
            try {
                const response = await axios.get(`${API_BASE}/repos/${this.statsForm.repo_id}/branches`);
                if (response.data.code === 0) {
                    this.statsFormBranches = response.data.data.branches || ['main'];
                    // 如果当前分支不在列表中，添加进去
                    if (this.statsForm.branch && !this.statsFormBranches.includes(this.statsForm.branch)) {
                        this.statsFormBranches.unshift(this.statsForm.branch);
                    }
                } else {
                    this.statsFormBranches = ['main', 'master', 'develop'];
                }
            } catch (error) {
                this.statsFormBranches = ['main', 'master', 'develop'];
            } finally {
                this.statsFormBranchesLoading = false;
            }
        },
        getRepoDisplayName(repo) {
            const url = repo.url;
            const parts = url.split('/');
            return parts[parts.length - 1].replace('.git', '');
        },
        async addRepos() {
            const urls = this.addRepoForm.urls.split('\n')
                .map(u => u.trim())
                .filter(u => u);
            
            if (urls.length === 0) {
                ElMessage.warning('请输入至少一个仓库URL');
                return;
            }

            const repos = urls.map(url => ({
                url,
                branch: this.addRepoForm.branch || 'main'
            }));

            const requestData = { repos };
            
            // 如果提供了认证信息
            if (this.addRepoForm.username && this.addRepoForm.password) {
                requestData.username = this.addRepoForm.username;
                requestData.password = this.addRepoForm.password;
            }

            try {
                const response = await axios.post(`${API_BASE}/repos/batch`, requestData);
                if (response.data.code === 0) {
                    const result = response.data.data;
                    ElMessage.success(`成功添加 ${result.success_count} 个仓库`);
                    if (result.failure_count > 0) {
                        ElMessage.warning(`${result.failure_count} 个仓库添加失败`);
                    }
                    this.addRepoVisible = false;
                    this.loadRepos();
                } else {
                    ElMessage.error(response.data.message || '添加仓库失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            }
        },
        async switchBranch(repo) {
            this.switchBranchForm = {
                branch: '',
                repoId: repo.id,
                repoUrl: repo.url,
                currentBranch: repo.current_branch || '未知'
            };
            this.branches = [];
            this.switchBranchVisible = true;
            
            // 获取分支列表
            await this.loadBranches(repo.id);
        },
        async loadBranches(repoId) {
            this.branchesLoading = true;
            try {
                const response = await axios.get(`${API_BASE}/repos/${repoId}/branches`);
                if (response.data.code === 0) {
                    this.branches = response.data.data.branches || [];
                } else {
                    ElMessage.warning('获取分支列表失败: ' + (response.data.message || ''));
                    this.branches = [];
                }
            } catch (error) {
                console.warn('无法获取分支列表:', error.message);
                this.branches = [];
            } finally {
                this.branchesLoading = false;
            }
        },
        async confirmSwitchBranch() {
            if (!this.switchBranchForm.branch) {
                ElMessage.warning('请输入分支名称');
                return;
            }

            try {
                const response = await axios.post(
                    `${API_BASE}/repos/${this.switchBranchForm.repoId}/switch-branch`,
                    { branch: this.switchBranchForm.branch }
                );
                if (response.data.code === 0) {
                    ElMessage.success('切换分支任务已提交');
                    this.switchBranchVisible = false;
                    this.loadRepos();
                } else {
                    ElMessage.error(response.data.message || '切换分支失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            }
        },
        async updateRepo(repoId) {
            try {
                const response = await axios.post(`${API_BASE}/repos/${repoId}/update`);
                if (response.data.code === 0) {
                    ElMessage.success('更新任务已提交');
                    this.loadRepos();
                } else {
                    ElMessage.error(response.data.message || '更新仓库失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            }
        },
        async resetRepo(repoId) {
            try {
                await ElMessageBox.confirm('确定要重置该仓库吗？', '警告', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                });

                const response = await axios.post(`${API_BASE}/repos/${repoId}/reset`);
                if (response.data.code === 0) {
                    ElMessage.success('重置任务已提交');
                    this.loadRepos();
                } else {
                    ElMessage.error(response.data.message || '重置仓库失败');
                }
            } catch (error) {
                if (error !== 'cancel') {
                    ElMessage.error('网络请求失败: ' + error.message);
                }
            }
        },
        async deleteRepo(repoId) {
            try {
                await ElMessageBox.confirm('确定要删除该仓库吗？此操作不可恢复！', '警告', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                });

                const response = await axios.delete(`${API_BASE}/repos/${repoId}`);
                if (response.data.code === 0) {
                    ElMessage.success('删除成功');
                    this.loadRepos();
                } else {
                    ElMessage.error(response.data.message || '删除仓库失败');
                }
            } catch (error) {
                if (error !== 'cancel') {
                    ElMessage.error('网络请求失败: ' + error.message);
                }
            }
        },
        async calculateStats() {
            if (!this.statsForm.repo_id) {
                ElMessage.warning('请选择仓库');
                return;
            }

            if (!this.statsForm.branch) {
                ElMessage.warning('请输入分支名称');
                return;
            }

            const constraint = {
                type: this.statsForm.constraint_type
            };

            if (this.statsForm.constraint_type === 'date_range') {
                if (!this.statsDateRange || this.statsDateRange.length !== 2) {
                    ElMessage.warning('请选择日期范围');
                    return;
                }
                constraint.from = this.statsDateRange[0];
                constraint.to = this.statsDateRange[1];
            } else {
                constraint.limit = this.statsForm.limit;
            }

            try {
                const response = await axios.post(`${API_BASE}/stats/calculate`, {
                    repo_id: this.statsForm.repo_id,
                    branch: this.statsForm.branch,
                    constraint
                });

                if (response.data.code === 0) {
                    ElMessage.success('统计任务已提交，请稍后查看结果');
                    // 3秒后刷新缓存列表
                    setTimeout(() => {
                        this.loadCaches();
                    }, 3000);
                } else {
                    ElMessage.error(response.data.message || '提交统计任务失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            }
        },
        async loadTasks() {
            this.tasksLoading = true;
            try {
                const response = await axios.get(`${API_BASE}/tasks`);
                if (response.data.code === 0) {
                    this.tasks = response.data.data.tasks || [];
                } else {
                    ElMessage.error(response.data.message || '加载任务列表失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            } finally {
                this.tasksLoading = false;
            }
        },
        async loadCaches() {
            this.cachesLoading = true;
            try {
                const response = await axios.get(`${API_BASE}/stats/caches`);
                if (response.data.code === 0) {
                    this.caches = response.data.data.caches || [];
                } else {
                    ElMessage.error(response.data.message || '加载统计缓存列表失败');
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
            } finally {
                this.cachesLoading = false;
            }
        },
        async clearAllCaches() {
            try {
                await ElMessageBox.confirm('确定要清空所有统计缓存吗？此操作不可恢复！', '警告', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                });

                const response = await axios.delete(`${API_BASE}/stats/caches/clear`);
                if (response.data.code === 0) {
                    ElMessage.success('所有统计缓存已清除');
                    this.loadCaches();
                } else {
                    ElMessage.error(response.data.message || '清除缓存失败');
                }
            } catch (error) {
                if (error !== 'cancel') {
                    ElMessage.error('网络请求失败: ' + error.message);
                }
            }
        },
        async clearAllTasks() {
            try {
                await ElMessageBox.confirm('确定要清空所有任务记录吗？包括正在执行的任务！', '警告', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'warning'
                });

                const response = await axios.delete(`${API_BASE}/tasks/clear`);
                if (response.data.code === 0) {
                    ElMessage.success('所有任务记录已清除');
                    this.loadTasks();
                } else {
                    ElMessage.error(response.data.message || '清除任务失败');
                }
            } catch (error) {
                if (error !== 'cancel') {
                    ElMessage.error('网络请求失败: ' + error.message);
                }
            }
        },
        async clearCompletedTasks() {
            try {
                await ElMessageBox.confirm('确定要清除所有已完成的任务记录吗？', '提示', {
                    confirmButtonText: '确定',
                    cancelButtonText: '取消',
                    type: 'info'
                });

                const response = await axios.delete(`${API_BASE}/tasks/clear-completed`);
                if (response.data.code === 0) {
                    ElMessage.success('已完成的任务记录已清除');
                    this.loadTasks();
                } else {
                    ElMessage.error(response.data.message || '清除已完成任务失败');
                }
            } catch (error) {
                if (error !== 'cancel') {
                    ElMessage.error('网络请求失败: ' + error.message);
                }
            }
        },
        async viewStatsCache(cache) {
            this.statsLoading = true;
            try {
                const params = {
                    repo_id: cache.repo_id,
                    branch: cache.branch,
                    constraint_type: cache.constraint_type
                };

                // 根据缓存的constraint_value添加参数
                if (cache.constraint_type === 'date_range' && cache.constraint_value) {
                    try {
                        const constraint = JSON.parse(cache.constraint_value);
                        if (constraint.from) params.from = constraint.from;
                        if (constraint.to) params.to = constraint.to;
                    } catch (e) {
                        console.error('Failed to parse constraint_value:', e);
                    }
                } else if (cache.constraint_type === 'commit_limit' && cache.constraint_value) {
                    try {
                        const constraint = JSON.parse(cache.constraint_value);
                        if (constraint.limit) params.limit = constraint.limit;
                    } catch (e) {
                        params.limit = 100; // 默认值
                    }
                }

                const response = await axios.get(`${API_BASE}/stats/result`, { params });

                if (response.data.code === 0) {
                    // 适配后端返回的数据结构
                    const data = response.data.data;
                    const stats = data.statistics;
                    
                    this.selectedStatsResult = {
                        summary: {
                            total_commits: stats.summary.total_commits || 0,
                            total_contributors: stats.summary.total_contributors || 0,
                            total_additions: stats.by_contributor.reduce((sum, c) => sum + (c.additions || 0), 0),
                            total_deletions: stats.by_contributor.reduce((sum, c) => sum + (c.deletions || 0), 0)
                        },
                        date_range: stats.summary.date_range || { from: '未指定', to: '未指定' },
                        contributors: stats.by_contributor.map(c => ({
                            name: c.author,
                            email: c.email,
                            commit_count: c.commits,
                            additions: c.additions,
                            deletions: c.deletions,
                            first_commit_date: c.first_commit_date || '-',
                            last_commit_date: c.last_commit_date || '-'
                        }))
                    };
                    ElMessage.success('查看统计结果成功');
                } else {
                    ElMessage.error(response.data.message || '查询统计结果失败');
                    this.selectedStatsResult = null;
                }
            } catch (error) {
                ElMessage.error('网络请求失败: ' + error.message);
                this.selectedStatsResult = null;
            } finally {
                this.statsLoading = false;
            }
        },
        getRepoStatusType(status) {
            const statusMap = {
                'pending': 'info',
                'cloning': 'warning',
                'ready': 'success',
                'error': 'danger'
            };
            return statusMap[status] || 'info';
        },
        getTaskStatusType(status) {
            const statusMap = {
                'pending': 'info',
                'running': 'warning',
                'completed': 'success',
                'failed': 'danger'
            };
            return statusMap[status] || 'info';
        },
        getConstraintText(cache) {
            if (!cache.constraint_value) return '未指定';
            try {
                const constraint = JSON.parse(cache.constraint_value);
                if (cache.constraint_type === 'date_range') {
                    return `${constraint.from || ''} ~ ${constraint.to || ''}`;
                } else if (cache.constraint_type === 'commit_limit') {
                    return `最近 ${constraint.limit || 100} 次提交`;
                }
            } catch (e) {
                return '解析失败';
            }
            return '未知';
        },
        getRepoName(repoId) {
            const repo = this.repos.find(r => r.id === repoId);
            if (repo) {
                const url = repo.url;
                const parts = url.split('/');
                return parts[parts.length - 1].replace('.git', '');
            }
            return `仓库 #${repoId}`;
        },
        formatDate(dateStr) {
            if (!dateStr) return '-';
            const date = new Date(dateStr);
            return date.toLocaleString('zh-CN');
        },
        formatFileSize(bytes) {
            if (!bytes) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
        }
    }
}).use(ElementPlus).mount('#app');
