// 从全局配置获取根路径，如果没有则默认为 "/"
const ROOT_PATH = (typeof window !== 'undefined' && window.ROOT_PATH) || '/';
const API_BASE = ROOT_PATH === '/' ? '/api' : ROOT_PATH + '/api';
let files = [];
let sortField = 'name'; // 当前排序字段: name, size, time
let sortOrder = 'asc';  // 排序方向: asc, desc
let currentPath = '';   // 当前路径

// DOM 元素
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');
const uploadBtn = document.getElementById('uploadBtn');
const refreshBtn = document.getElementById('refreshBtn');
const filesContainer = document.getElementById('filesContainer');
const breadcrumb = document.getElementById('breadcrumb');
const toast = document.getElementById('toast');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    loadFiles();
    updateSortIcons();
});

// 设置事件监听器
function setupEventListeners() {
    // 上传按钮点击
    uploadBtn.addEventListener('click', () => {
        fileInput.click();
    });

    // 文件选择
    fileInput.addEventListener('change', (e) => {
        handleFiles(e.target.files);
    });

    // 拖拽上传
    uploadArea.addEventListener('click', () => {
        fileInput.click();
    });

    uploadArea.addEventListener('dragover', (e) => {
        e.preventDefault();
        uploadArea.classList.add('dragover');
    });

    uploadArea.addEventListener('dragleave', () => {
        uploadArea.classList.remove('dragover');
    });

    uploadArea.addEventListener('drop', (e) => {
        e.preventDefault();
        uploadArea.classList.remove('dragover');
        handleFiles(e.dataTransfer.files);
    });

    // 刷新按钮
    refreshBtn.addEventListener('click', () => {
        loadFiles();
    });

    // 排序按钮
    document.querySelectorAll('.sortable').forEach(th => {
        th.addEventListener('click', () => {
            const field = th.dataset.sort;
            if (sortField === field) {
                // 切换排序方向
                sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                // 新的排序字段，默认升序
                sortField = field;
                sortOrder = 'asc';
            }
            sortFiles();
            renderFiles();
            updateSortIcons();
        });
    });
}

// 处理文件上传
function handleFiles(fileList) {
    if (fileList.length === 0) return;

    Array.from(fileList).forEach(file => {
        uploadFile(file);
    });
    
    // 清空文件选择，允许重复选择同一文件
    fileInput.value = '';
}

// 上传单个文件
function uploadFile(file) {
    const formData = new FormData();
    formData.append('file', file);

    // 创建进度条
    const progressId = 'progress-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    const progressContainer = document.getElementById('uploadProgress');
    progressContainer.style.display = 'block';
    
    const progressItem = document.createElement('div');
    progressItem.className = 'upload-progress-item';
    progressItem.id = progressId;
    progressItem.innerHTML = `
        <div class="progress-header">
            <span class="progress-filename">${file.name}</span>
            <span class="progress-percent">0%</span>
        </div>
        <div class="progress-bar">
            <div class="progress-bar-fill" style="width: 0%"></div>
        </div>
    `;
    progressContainer.appendChild(progressItem);

    const xhr = new XMLHttpRequest();
    const progressBar = progressItem.querySelector('.progress-bar-fill');
    const progressPercent = progressItem.querySelector('.progress-percent');

    xhr.upload.addEventListener('progress', (e) => {
        if (e.lengthComputable) {
            const percent = Math.round((e.loaded / e.total) * 100);
            progressBar.style.width = percent + '%';
            progressPercent.textContent = percent + '%';
        }
    });

    xhr.addEventListener('load', () => {
        if (xhr.status === 200) {
            try {
                const data = JSON.parse(xhr.responseText);
                if (data.success) {
                    progressItem.classList.add('success');
                    progressPercent.textContent = '完成';
                    setTimeout(() => {
                        progressItem.remove();
                        if (progressContainer.children.length === 0) {
                            progressContainer.style.display = 'none';
                        }
                    }, 1000);
                    // 延迟刷新文件列表，避免多个文件同时上传时频繁刷新
                    setTimeout(() => {
                        loadFiles(currentPath);
                    }, 500);
                } else {
                    progressItem.classList.add('error');
                    progressPercent.textContent = '失败';
                    showToast(data.message || '上传失败', 'error');
                }
            } catch (e) {
                progressItem.classList.add('error');
                progressPercent.textContent = '失败';
                showToast('上传失败', 'error');
            }
        } else {
            progressItem.classList.add('error');
            progressPercent.textContent = '失败';
            showToast('上传失败: HTTP ' + xhr.status, 'error');
        }
    });

    xhr.addEventListener('error', () => {
        progressItem.classList.add('error');
        progressPercent.textContent = '失败';
        showToast('上传失败: 网络错误', 'error');
    });

    xhr.addEventListener('abort', () => {
        progressItem.remove();
        if (progressContainer.children.length === 0) {
            progressContainer.style.display = 'none';
        }
    });

    // 构建上传URL，包含当前路径
    let uploadUrl = `${API_BASE}/upload`;
    if (currentPath) {
        uploadUrl += `?path=${encodeURIComponent(currentPath)}`;
    }
    
    xhr.open('POST', uploadUrl);
    xhr.send(formData);
}

// 加载文件列表
async function loadFiles(path = '') {
    try {
        currentPath = path;
        filesContainer.innerHTML = '<tr><td colspan="5" class="loading">加载中...</td></tr>';
        
        const url = path ? `${API_BASE}/files?path=${encodeURIComponent(path)}` : `${API_BASE}/files`;
        const response = await fetch(url);
        const data = await response.json();

        if (data.success) {
            files = data.data || [];
            sortFiles();
            renderFiles();
            updateSortIcons();
            updateBreadcrumb(path);
        } else {
            showToast('加载文件列表失败', 'error');
            filesContainer.innerHTML = `
                <tr>
                    <td colspan="5" class="empty-state">加载失败</td>
                </tr>
            `;
        }
    } catch (error) {
        showToast('加载文件列表失败: ' + error.message, 'error');
        filesContainer.innerHTML = `
            <tr>
                <td colspan="5" class="empty-state">加载失败</td>
            </tr>
        `;
    }
}

// 更新面包屑导航
function updateBreadcrumb(path) {
    if (!path) {
        breadcrumb.innerHTML = '<span class="breadcrumb-item" data-path="">根目录</span>';
        return;
    }
    
    const parts = path.split(/[/\\]/).filter(p => p);
    let html = '<span class="breadcrumb-item" data-path="">根目录</span>';
    
    let current = '';
    parts.forEach((part, index) => {
        current = current ? current + '/' + part : part;
        html += ` <span class="breadcrumb-separator">/</span> <span class="breadcrumb-item" data-path="${current}">${part}</span>`;
    });
    
    breadcrumb.innerHTML = html;
    
    // 添加点击事件
    breadcrumb.querySelectorAll('.breadcrumb-item').forEach(item => {
        item.addEventListener('click', () => {
            const targetPath = item.dataset.path || '';
            loadFiles(targetPath);
        });
    });
}

// 进入目录
function enterDirectory(path) {
    loadFiles(path);
}

// 渲染文件列表
function renderFiles() {
    if (files.length === 0) {
        filesContainer.innerHTML = `
            <tr>
                <td colspan="5" class="empty-state">
                    <div class="empty-state-icon">📂</div>
                    <p>暂无文件</p>
                    <p style="margin-top: 10px; font-size: 0.9em;">上传您的第一个文件开始使用</p>
                </td>
            </tr>
        `;
        return;
    }

    filesContainer.innerHTML = files.map(file => createFileRow(file)).join('');
    
    // 添加目录点击事件
    document.querySelectorAll('.file-dir').forEach(item => {
        item.addEventListener('click', (e) => {
            const path = e.currentTarget.dataset.path;
            enterDirectory(path);
        });
    });
    
    // 添加事件监听器
    document.querySelectorAll('.btn-download').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const path = e.target.dataset.path;
            downloadFile(path);
        });
    });

    document.querySelectorAll('.btn-danger').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const path = e.target.dataset.path;
            deleteFile(path);
        });
    });
}

// 创建文件表格行
function createFileRow(file) {
    const icon = file.isDir ? '📁' : getFileIcon(file.extension);
    const size = file.isDir ? '-' : formatFileSize(file.size);
    const date = formatDate(file.modTime);
    const rowClass = file.isDir ? 'file-dir' : '';
    const path = file.path || file.name;

    return `
        <tr class="${rowClass}" data-path="${path}">
            <td>${icon}</td>
            <td title="${file.name}" class="${file.isDir ? 'dir-name' : ''}">${file.name}${file.isDir ? ' /' : ''}</td>
            <td>${size}</td>
            <td>${date}</td>
            <td>
                <div class="file-actions">
                    ${file.isDir ? '' : `<button class="btn btn-download" data-path="${path}">下载</button>`}
                    <button class="btn btn-danger" data-path="${path}">删除</button>
                </div>
            </td>
        </tr>
    `;
}

// 获取文件图标
function getFileIcon(extension) {
    const icons = {
        'pdf': '📄',
        'doc': '📝', 'docx': '📝',
        'xls': '📊', 'xlsx': '📊',
        'ppt': '📽️', 'pptx': '📽️',
        'jpg': '🖼️', 'jpeg': '🖼️', 'png': '🖼️', 'gif': '🖼️', 'svg': '🖼️',
        'mp4': '🎬', 'avi': '🎬', 'mov': '🎬',
        'mp3': '🎵', 'wav': '🎵',
        'zip': '📦', 'rar': '📦', '7z': '📦',
        'txt': '📃',
        'js': '📜', 'html': '📜', 'css': '📜',
        'exe': '⚙️',
    };
    return icons[extension?.toLowerCase()] || '📄';
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// 格式化日期
function formatDate(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
        return '今天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else if (days === 1) {
        return '昨天 ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    } else if (days < 7) {
        return days + ' 天前';
    } else {
        return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    }
}

// 下载文件
function downloadFile(path) {
    window.open(`${API_BASE}/download/${encodeURIComponent(path)}`, '_blank');
    showToast('开始下载...', 'success');
}

// 删除文件
async function deleteFile(path) {
    const pathParts = path.split(/[/\\]/);
    const name = pathParts[pathParts.length - 1];
    
    if (!confirm(`确定要删除 "${name}" 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/delete/${encodeURIComponent(path)}`, {
            method: 'DELETE'
        });

        const data = await response.json();

        if (data.success) {
            showToast(data.message || '删除成功', 'success');
            loadFiles(currentPath);
        } else {
            showToast(data.message || '删除失败', 'error');
        }
    } catch (error) {
        showToast('删除失败: ' + error.message, 'error');
    }
}

// 排序文件
function sortFiles() {
    files.sort((a, b) => {
        // 目录始终排在前面
        if (a.isDir && !b.isDir) return -1;
        if (!a.isDir && b.isDir) return 1;
        
        let compareA, compareB;
        
        switch (sortField) {
            case 'name':
                compareA = a.name.toLowerCase();
                compareB = b.name.toLowerCase();
                break;
            case 'size':
                compareA = a.size;
                compareB = b.size;
                break;
            case 'time':
                compareA = new Date(a.modTime).getTime();
                compareB = new Date(b.modTime).getTime();
                break;
            default:
                return 0;
        }
        
        if (compareA < compareB) {
            return sortOrder === 'asc' ? -1 : 1;
        }
        if (compareA > compareB) {
            return sortOrder === 'asc' ? 1 : -1;
        }
        return 0;
    });
}

// 更新排序图标
function updateSortIcons() {
    document.querySelectorAll('.sortable').forEach(th => {
        const icon = th.querySelector('.sort-icon');
        const field = th.dataset.sort;
        
        if (sortField === field) {
            icon.textContent = sortOrder === 'asc' ? ' ↑' : ' ↓';
            icon.style.opacity = '1';
        } else {
            icon.textContent = '';
            icon.style.opacity = '0';
        }
    });
}

// 显示提示消息
function showToast(message, type = 'info') {
    toast.textContent = message;
    toast.className = `toast ${type} show`;

    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

