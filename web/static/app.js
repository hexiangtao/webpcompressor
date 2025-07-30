class WebPCompressor {
    constructor() {
        this.currentTask = null;
        this.websocket = null;
        this.uploadedFile = null;
        
        this.initializeElements();
        this.bindEvents();
        this.loadSystemInfo();
        this.loadTaskHistory();
    }

    initializeElements() {
        // 主要元素
        this.uploadZone = document.getElementById('uploadZone');
        this.fileInput = document.getElementById('fileInput');
        this.settingsPanel = document.getElementById('settingsPanel');
        this.progressPanel = document.getElementById('progressPanel');
        this.resultPanel = document.getElementById('resultPanel');
        
        // 设置元素
        this.qualitySlider = document.getElementById('qualitySlider');
        this.qualityValue = document.getElementById('qualityValue');
        this.presetSelect = document.getElementById('presetSelect');
        this.losslessCheck = document.getElementById('losslessCheck');
        this.parallelCheck = document.getElementById('parallelCheck');
        
        // 按钮
        this.selectBtn = document.getElementById('selectBtn');
        this.compressBtn = document.getElementById('compressBtn');
        this.cancelBtn = document.getElementById('cancelBtn');
        this.downloadBtn = document.getElementById('downloadBtn');
        this.newTaskBtn = document.getElementById('newTaskBtn');
        
        // 进度元素
        this.progressBar = document.getElementById('progressBar');
        this.progressPercent = document.getElementById('progressPercent');
        this.progressMessage = document.getElementById('progressMessage');
        
        // 结果元素
        this.originalSize = document.getElementById('originalSize');
        this.compressedSize = document.getElementById('compressedSize');
        this.compressionRatio = document.getElementById('compressionRatio');
        
        // 模态框
        this.statsModal = document.getElementById('statsModal');
        this.helpModal = document.getElementById('helpModal');
        this.statsContent = document.getElementById('statsContent');
        this.taskHistory = document.getElementById('taskHistory');
    }

    bindEvents() {
        // 文件上传事件
        this.uploadZone.addEventListener('click', () => this.fileInput.click());
        this.fileInput.addEventListener('change', (e) => this.handleFileSelect(e));
        
        // 拖拽事件
        this.uploadZone.addEventListener('dragover', (e) => this.handleDragOver(e));
        this.uploadZone.addEventListener('dragleave', (e) => this.handleDragLeave(e));
        this.uploadZone.addEventListener('drop', (e) => this.handleFileDrop(e));
        
        // 设置事件
        this.qualitySlider.addEventListener('input', (e) => {
            this.qualityValue.textContent = e.target.value;
        });
        
        // 按钮事件
        this.compressBtn.addEventListener('click', () => this.startCompression());
        this.cancelBtn.addEventListener('click', () => this.cancelTask());
        this.downloadBtn.addEventListener('click', () => this.downloadFile());
        this.newTaskBtn.addEventListener('click', () => this.resetInterface());
        
        // 模态框事件
        document.getElementById('statsBtn').addEventListener('click', () => this.showStats());
        document.getElementById('helpBtn').addEventListener('click', () => this.showHelp());
        document.getElementById('closeStatsBtn').addEventListener('click', () => this.hideStats());
        document.getElementById('closeHelpBtn').addEventListener('click', () => this.hideHelp());
    }

    handleDragOver(e) {
        e.preventDefault();
        this.uploadZone.classList.add('drag-over');
    }

    handleDragLeave(e) {
        e.preventDefault();
        this.uploadZone.classList.remove('drag-over');
    }

    handleFileDrop(e) {
        e.preventDefault();
        this.uploadZone.classList.remove('drag-over');
        
        const files = e.dataTransfer.files;
        if (files.length > 0) {
            this.handleFile(files[0]);
        }
    }

    handleFileSelect(e) {
        const file = e.target.files[0];
        if (file) {
            this.handleFile(file);
        }
    }

    async handleFile(file) {
        // 验证文件类型
        if (!file.name.toLowerCase().endsWith('.webp')) {
            this.showNotification('请选择WebP格式的文件', 'error');
            return;
        }

        // 验证文件大小
        if (file.size > 100 * 1024 * 1024) { // 100MB
            this.showNotification('文件大小不能超过100MB', 'error');
            return;
        }

        try {
            // 上传文件
            this.showNotification('正在上传文件...', 'info');
            const uploadResult = await this.uploadFile(file);
            
            this.uploadedFile = uploadResult.file;
            this.showNotification('文件上传成功！', 'success');
            
            // 显示设置面板
            this.settingsPanel.classList.remove('hidden');
            
            // 更新界面
            this.updateUploadZone(file);
            
        } catch (error) {
            this.showNotification('文件上传失败: ' + error.message, 'error');
        }
    }

    async uploadFile(file) {
        const formData = new FormData();
        formData.append('file', file);

        console.log('开始上传文件:', file.name, file.size);

        try {
            const response = await fetch('/api/v1/upload', {
                method: 'POST',
                body: formData
            });

            console.log('上传响应状态:', response.status);

            if (!response.ok) {
                const error = await response.json();
                console.error('上传失败:', error);
                throw new Error(error.error || '上传失败');
            }

            const result = await response.json();
            console.log('上传成功:', result);
            return result;
        } catch (error) {
            console.error('上传出错:', error);
            throw error;
        }
    }

    updateUploadZone(file) {
        this.uploadZone.innerHTML = `
            <i class="fas fa-file-image text-6xl text-green-500 mb-4"></i>
            <h3 class="text-2xl font-semibold text-gray-700 mb-2">${file.name}</h3>
            <p class="text-gray-500 mb-4">文件大小: ${this.formatFileSize(file.size)}</p>
            <button class="bg-gray-500 hover:bg-gray-600 text-white px-6 py-2 rounded-lg font-medium">
                <i class="fas fa-redo mr-2"></i>重新选择
            </button>
        `;
    }

    async startCompression() {
        if (!this.uploadedFile) {
            this.showNotification('请先上传文件', 'error');
            return;
        }

        try {
            // 创建压缩任务
            const taskData = {
                input_file: this.uploadedFile.filename,
                quality: parseInt(this.qualitySlider.value),
                preset: this.presetSelect.value,
                lossless: this.losslessCheck.checked,
                parallel: this.parallelCheck.checked
            };

            console.log('创建任务数据:', taskData);
            console.log('上传文件信息:', this.uploadedFile);

            const response = await fetch('/api/v1/tasks', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(taskData)
            });

            console.log('任务创建响应状态:', response.status);
            
            if (!response.ok) {
                const error = await response.json();
                console.error('任务创建失败:', error);
                throw new Error(error.error || '创建任务失败');
            }

            const result = await response.json();
            this.currentTask = result.task;

            // 隐藏设置面板，显示进度面板
            this.settingsPanel.classList.add('hidden');
            this.progressPanel.classList.remove('hidden');

            // 连接WebSocket监听进度
            this.connectWebSocket(this.currentTask.id);

            this.showNotification('压缩任务已启动', 'success');

        } catch (error) {
            this.showNotification('启动压缩失败: ' + error.message, 'error');
        }
    }

    connectWebSocket(taskId) {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/v1/progress/${taskId}`;
        
        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
            console.log('WebSocket连接已建立');
        };

        this.websocket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleProgressUpdate(data);
        };

        this.websocket.onclose = () => {
            console.log('WebSocket连接已关闭');
        };

        this.websocket.onerror = (error) => {
            console.error('WebSocket错误:', error);
            this.showNotification('连接出现问题，请刷新页面重试', 'error');
        };
    }

    handleProgressUpdate(data) {
        if (data.type === 'progress_update') {
            this.updateProgress(data.progress, data.message);
        } else if (data.type === 'task_update') {
            if (data.status === 'completed') {
                this.handleTaskComplete(data);
            } else if (data.status === 'failed') {
                this.handleTaskFailed(data);
            }
        }
    }

    updateProgress(progress, message) {
        this.progressBar.style.width = progress + '%';
        this.progressPercent.textContent = Math.round(progress) + '%';
        this.progressMessage.textContent = message;
    }

    async handleTaskComplete(data) {
        // 关闭WebSocket
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }

        // 隐藏进度面板，显示结果面板
        this.progressPanel.classList.add('hidden');
        this.resultPanel.classList.remove('hidden');

        // 更新结果信息
        if (data.result) {
            this.originalSize.textContent = this.formatFileSize(data.result.original_size);
            this.compressedSize.textContent = this.formatFileSize(data.result.compressed_size);
            this.compressionRatio.textContent = data.result.compression_ratio.toFixed(1) + '%';
        }

        // 设置下载链接
        this.downloadBtn.onclick = () => {
            const outputFile = this.currentTask.output_file.split('/').pop();
            window.open(`/api/v1/download/${outputFile}`, '_blank');
        };

        this.showNotification('压缩完成！', 'success');
        this.loadTaskHistory(); // 刷新任务历史
    }

    handleTaskFailed(data) {
        // 关闭WebSocket
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }

        // 隐藏进度面板
        this.progressPanel.classList.add('hidden');

        this.showNotification('压缩失败: ' + (data.error || '未知错误'), 'error');
        this.resetInterface();
    }

    async cancelTask() {
        if (!this.currentTask) return;

        try {
            const response = await fetch(`/api/v1/tasks/${this.currentTask.id}/cancel`, {
                method: 'POST'
            });

            if (response.ok) {
                this.showNotification('任务已取消', 'info');
                this.resetInterface();
            }
        } catch (error) {
            this.showNotification('取消任务失败', 'error');
        }
    }

    resetInterface() {
        // 关闭WebSocket
        if (this.websocket) {
            this.websocket.close();
            this.websocket = null;
        }

        // 隐藏所有面板
        this.settingsPanel.classList.add('hidden');
        this.progressPanel.classList.add('hidden');
        this.resultPanel.classList.add('hidden');

        // 重置上传区域
        this.uploadZone.innerHTML = `
            <i class="fas fa-cloud-upload-alt text-6xl text-gray-400 mb-4"></i>
            <h3 class="text-2xl font-semibold text-gray-700 mb-2">拖拽文件到这里</h3>
            <p class="text-gray-500 mb-4">或者点击选择WebP文件</p>
            <button id="selectBtn" class="bg-blue-500 hover:bg-blue-600 text-white px-6 py-3 rounded-lg font-medium">
                <i class="fas fa-file-upload mr-2"></i>选择文件
            </button>
        `;

        // 重新绑定选择按钮事件
        document.getElementById('selectBtn').addEventListener('click', () => this.fileInput.click());

        // 重置状态
        this.currentTask = null;
        this.uploadedFile = null;
        this.fileInput.value = '';
    }

    async loadSystemInfo() {
        try {
            const response = await fetch('/api/v1/info');
            const info = await response.json();
            
            // 可以在这里使用系统信息更新界面
            console.log('系统信息:', info);
        } catch (error) {
            console.error('加载系统信息失败:', error);
        }
    }

    async loadTaskHistory() {
        try {
            const response = await fetch('/api/v1/tasks?limit=10');
            const data = await response.json();
            
            this.updateTaskHistory(data.tasks || []);
        } catch (error) {
            console.error('加载任务历史失败:', error);
        }
    }

    updateTaskHistory(tasks) {
        if (tasks.length === 0) {
            this.taskHistory.innerHTML = '<p class="text-gray-500 text-center py-8">暂无任务历史</p>';
            return;
        }

        const tasksHtml = tasks.map(task => {
            const statusIcon = this.getStatusIcon(task.status);
            const statusColor = this.getStatusColor(task.status);
            
            return `
                <div class="flex items-center justify-between p-4 border border-gray-200 rounded-lg">
                    <div class="flex items-center space-x-4">
                        <i class="${statusIcon} ${statusColor} text-xl"></i>
                        <div>
                            <div class="font-medium text-gray-800">${task.input_file.split('/').pop()}</div>
                            <div class="text-sm text-gray-500">
                                质量: ${task.config.quality}% | 
                                状态: ${this.getStatusText(task.status)} |
                                创建时间: ${new Date(task.created_at).toLocaleString()}
                            </div>
                        </div>
                    </div>
                    <div class="flex items-center space-x-2">
                        ${task.status === 'completed' && task.result ? 
                            `<span class="text-sm text-green-600">节省 ${this.formatFileSize(task.result.original_size - task.result.compressed_size)}</span>` : 
                            ''}
                        ${task.status === 'completed' ? 
                            `<button onclick="window.open('/api/v1/download/${task.output_file.split('/').pop()}', '_blank')" 
                                     class="text-blue-600 hover:text-blue-800">
                                <i class="fas fa-download"></i>
                             </button>` : ''}
                    </div>
                </div>
            `;
        }).join('');

        this.taskHistory.innerHTML = tasksHtml;
    }

    getStatusIcon(status) {
        const icons = {
            'pending': 'fas fa-clock',
            'processing': 'fas fa-spinner fa-spin',
            'completed': 'fas fa-check-circle',
            'failed': 'fas fa-exclamation-circle',
            'cancelled': 'fas fa-times-circle'
        };
        return icons[status] || 'fas fa-question-circle';
    }

    getStatusColor(status) {
        const colors = {
            'pending': 'text-yellow-500',
            'processing': 'text-blue-500',
            'completed': 'text-green-500',
            'failed': 'text-red-500',
            'cancelled': 'text-gray-500'
        };
        return colors[status] || 'text-gray-500';
    }

    getStatusText(status) {
        const texts = {
            'pending': '待处理',
            'processing': '处理中',
            'completed': '已完成',
            'failed': '失败',
            'cancelled': '已取消'
        };
        return texts[status] || '未知';
    }

    async showStats() {
        try {
            const response = await fetch('/api/v1/stats');
            const stats = await response.json();
            
            this.statsContent.innerHTML = `
                <div class="grid grid-cols-2 gap-4">
                    <div class="text-center">
                        <div class="text-2xl font-bold text-blue-600">${stats.total_tasks || 0}</div>
                        <div class="text-sm text-gray-500">总任务数</div>
                    </div>
                    <div class="text-center">
                        <div class="text-2xl font-bold text-green-600">${stats.completed_tasks || 0}</div>
                        <div class="text-sm text-gray-500">已完成</div>
                    </div>
                    <div class="text-center">
                        <div class="text-2xl font-bold text-red-600">${stats.failed_tasks || 0}</div>
                        <div class="text-sm text-gray-500">失败</div>
                    </div>
                    <div class="text-center">
                        <div class="text-2xl font-bold text-yellow-600">${stats.processing_tasks || 0}</div>
                        <div class="text-sm text-gray-500">处理中</div>
                    </div>
                </div>
                ${stats.total_saved_bytes ? `
                    <div class="text-center mt-4 pt-4 border-t">
                        <div class="text-lg font-semibold text-purple-600">
                            总共节省: ${this.formatFileSize(stats.total_saved_bytes)}
                        </div>
                    </div>
                ` : ''}
            `;
            
            this.statsModal.classList.remove('hidden');
        } catch (error) {
            this.showNotification('加载统计信息失败', 'error');
        }
    }

    showHelp() {
        this.helpModal.classList.remove('hidden');
    }

    hideStats() {
        this.statsModal.classList.add('hidden');
    }

    hideHelp() {
        this.helpModal.classList.add('hidden');
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    showNotification(message, type = 'info') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg text-white z-50 ${
            type === 'success' ? 'bg-green-500' :
            type === 'error' ? 'bg-red-500' :
            type === 'warning' ? 'bg-yellow-500' : 'bg-blue-500'
        }`;
        notification.textContent = message;

        document.body.appendChild(notification);

        // 3秒后自动移除
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 3000);
    }
}

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    new WebPCompressor();
}); 