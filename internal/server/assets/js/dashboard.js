let socket;
    let reconnectAttempts = 0;
    const maxReconnectAttempts = 5;
    const reconnectDelay = 3000;

    function connectWebSocket() {
        socket = new WebSocket('ws://' + window.location.host + '/ws');

        socket.onopen = function() {
            console.log('WebSocket connected');
            reconnectAttempts = 0;
        };

        socket.onmessage = function(event) {
            const data = JSON.parse(event.data);
            updateDashboard(data);
        };

        socket.onclose = function() {
            console.log('WebSocket disconnected');
            if (reconnectAttempts < maxReconnectAttempts) {
                setTimeout(connectWebSocket, reconnectDelay);
                reconnectAttempts++;
            } else {
                console.error('Max reconnect attempts reached');
            }
        };
    }

    function fetchInitialData() {
        fetch('/initial-data')
            .then(response => response.json())
            .then(data => {
                updateDashboard(data);
                document.getElementById('loading').style.display = 'none';
                document.getElementById('dashboard').style.display = 'block';
            })
            .catch(error => console.error('Error fetching initial data:', error));
    }

    function updateDashboard(data) {
        const versionElement = document.getElementById('version');
        if (versionElement) {
            versionElement.textContent = data.Version;
            if (data.Version === "dev") {
                versionElement.textContent = "Development Version";
                versionElement.style.backgroundColor = "#dc3545";
            }
        }

        const container = document.getElementById('characters-container');
        if (!container) return;

        if (Object.keys(data.Status).length === 0) {
            container.innerHTML = '<article><p>No characters found, start adding a new character.</p></article>';
            return;
        }

        for (const [key, value] of Object.entries(data.Status)) {
            let card = document.getElementById(`card-${key}`);
            if (!card) {
                card = createCharacterCard(key);
                container.appendChild(card);
            }
            updateCharacterCard(card, key, value, data.DropCount[key]);
        }

        // Remove cards for characters that no longer exist
        Array.from(container.children).forEach(card => {
            if (!data.Status.hasOwnProperty(card.id.replace('card-', ''))) {
                container.removeChild(card);
            }
        });
    }


    function createCharacterCard(key) {
        const card = document.createElement('div');
        card.className = 'character-card';
        card.id = `card-${key}`;

        card.innerHTML = `
            <div class="character-header">
                <div class="character-name">
                    <span>${key}</span>
                     <div class="status-indicator"></div>
                </div>
                <div class="character-controls">
                    <button class="btn btn-outline" onclick="location.href='/debug?characterName=${key}'">
                        <i class="bi bi-bug btn-icon"></i>Debug
                    </button>
                    <button class="btn btn-outline" onclick="location.href='/supervisorSettings?supervisor=${key}'">
                        <i class="bi bi-gear btn-icon"></i>Settings
                    </button>
                    <button class="start-pause btn btn-start" data-character="${key}">
                        <i class="bi bi-play-fill btn-icon"></i>Start
                    </button>
                    <button class="stop btn btn-stop" data-character="${key}" style="display:none;">
                        <i class="bi bi-stop-fill btn-icon"></i>Stop
                    </button>
                    <button class="toggle-details">
                        <i class="bi bi-chevron-down"></i>
                    </button>
                </div>
            </div>
            <div class="character-details">
                <div class="status-details">
                    <span class="status-badge"></span>
                </div>
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-label">Games</div>
                        <div class="stat-value runs">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Drops</div>
                        <div class="stat-value drops">None</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Chickens</div>
                        <div class="stat-value chickens">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Deaths</div>
                        <div class="stat-value deaths">0</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-label">Errors</div>
                        <div class="stat-value errors">0</div>
                    </div>
                </div>
                <div class="run-stats"></div>
            </div>
        `;

        setupEventListeners(card, key);
        return card;
    }


    function setupEventListeners(card, key) {
        if (!card) return;

        const toggleDetailsBtn = card.querySelector('.toggle-details');
        const startPauseBtn = card.querySelector('.start-pause');
        const stopBtn = card.querySelector('.stop');

        if (toggleDetailsBtn) {
            toggleDetailsBtn.addEventListener('click', function() {
                card.classList.toggle('expanded');
                this.querySelector('i').style.transform = card.classList.contains('expanded') ? 'rotate(180deg)' : 'rotate(0deg)';
                saveExpandedState();
            });
        }

        if (startPauseBtn) {
            startPauseBtn.addEventListener('click', function() {
                const action = this.textContent.trim() === 'Start' ? 'start' : 'togglePause';
                fetch(`/${action}?characterName=${key}`).then(() => fetchInitialData());
            });
        }

        if (stopBtn) {
            stopBtn.addEventListener('click', function() {
                fetch(`/stop?characterName=${key}`).then(() => fetchInitialData());
            });
        }
    }


    function updateStatusPosition(card, isExpanded) {
        if (!card) return;

        const statusBadge = card.querySelector('.status-badge');
        const headerStatusContainer = card.querySelector('.character-name-status');
        const detailsStatusContainer = card.querySelector('.status-details');

        if (!statusBadge || !headerStatusContainer || !detailsStatusContainer) return;

        if (isExpanded) {
            detailsStatusContainer.insertBefore(statusBadge, detailsStatusContainer.firstChild);
        } else {
            headerStatusContainer.appendChild(statusBadge);
        }
    }

    function updateCharacterCard(card, key, value, dropCount) {
        if (!card) return;

        const startPauseBtn = card.querySelector('.start-pause');
        const stopBtn = card.querySelector('.stop');
        const statusDetails = card.querySelector('.status-details');
        const statusBadge = statusDetails.querySelector('.status-badge');
        const statusIndicator = card.querySelector('.status-indicator');

        if (statusBadge && statusDetails) {
            updateStatus(statusBadge, statusDetails, value.SupervisorStatus);
        }
        
        if (statusIndicator) {
                updateStatusIndicator(statusIndicator, value.SupervisorStatus);
        }

        if (startPauseBtn && stopBtn) {
            updateButtons(startPauseBtn, stopBtn, value.SupervisorStatus);
        }
        
        updateStats(card, key, value.Games, dropCount);
        updateRunStats(card, value.Games);
        
        if (statusDetails) {
            updateStartedTime(statusDetails, value.StartedAt);
        }
    }

    function updateStatusIndicator(statusIndicator, status) {
        statusIndicator.classList.remove('in-game', 'paused', 'stopped');
        if (status === "In game") {
            statusIndicator.classList.add('in-game');
        } else if (status === "Starting") {
            statusIndicator.classList.add('paused');
        } else if (status === "Paused") {
            statusIndicator.classList.add('paused');
        } else {
            statusIndicator.classList.add('stopped');
        }
    }

    function updateStatus(statusBadge, statusDetails, status) {
        if (!statusBadge || !statusDetails) return;

        const statusText = status || 'Not started';
        statusBadge.innerHTML = `<span class="status-label">Status:</span> <span class="status-value">${statusText}</span>`;
        statusBadge.className = `status-badge status-${statusText.toLowerCase().replace(' ', '')}`;
    }

    function updateStartedTime(statusDetails, startedAt) {
        const startTime = new Date(startedAt);
        const now = new Date();
        
        let runningForElement = statusDetails.querySelector('.running-for');
        if (!runningForElement) {
            runningForElement = document.createElement('div');
            runningForElement.className = 'running-for';
            statusDetails.appendChild(runningForElement);
        }
        
        if (startTime.getFullYear() === 1) {
            runningForElement.textContent = 'Running for: N/A';
            return;
        }
        
        const timeDiff = now - startTime;
        if (timeDiff < 0) {
            runningForElement.textContent = 'Running for: N/A';
            return;
        }
        
        const duration = formatDuration(timeDiff);
        runningForElement.textContent = `Running for: ${duration}`;
    }

    function updateButtons(startPauseBtn, stopBtn, status) {
        if (status === "Paused") {
            startPauseBtn.innerHTML = '<i class="bi bi-play-fill btn-icon"></i>Resume';
            startPauseBtn.className = 'start-pause btn btn-pause';
            stopBtn.style.display = 'inline-block';
        } else if (status === "In game" || status === "Starting") {
            startPauseBtn.innerHTML = '<i class="bi bi-pause-fill btn-icon"></i>Pause';
            startPauseBtn.className = 'start-pause btn btn-pause';
            stopBtn.style.display = 'inline-block';
        } else {
            startPauseBtn.innerHTML = '<i class="bi bi-play-fill btn-icon"></i>Start';
            startPauseBtn.className = 'start-pause btn btn-start';
            stopBtn.style.display = 'none';
        }
    }

    function updateStats(card, key, games, dropCount) {
        const stats = calculateStats(games);
        
        card.querySelector('.runs').textContent = stats.totalGames;
        card.querySelector('.drops').innerHTML = dropCount === undefined ? 'None' : 
            (dropCount === 0 ? 'None' : `<a href="/drops?supervisor=${key}">${dropCount}</a>`);
        card.querySelector('.chickens').textContent = stats.totalChickens;
        card.querySelector('.deaths').textContent = stats.totalDeaths;
        card.querySelector('.errors').textContent = stats.totalErrors;
    }


    function updateRunStats(card, games) {
    const runStats = calculateRunStats(games);
    const runStatsElement = card.querySelector('.run-stats');
    runStatsElement.innerHTML = '<h3>Run Statistics</h3>';

    if (Object.keys(runStats).length === 0) {
        runStatsElement.innerHTML += '<p>No run data available yet.</p>';
        return;
    }

    const runStatsGrid = document.createElement('div');
    runStatsGrid.className = 'run-stats-grid';

    for (const [runName, stats] of Object.entries(runStats)) {
        const runElement = document.createElement('div');
        runElement.className = 'run-stat';
        if (stats.isCurrentRun) {
            runElement.classList.add('current-run');
        }
        runElement.innerHTML = `
            <h4>${runName}${stats.isCurrentRun ? ' <span class="current-run-indicator">Current</span>' : ''}</h4>
            <div class="run-stat-content">
                <div class="run-stat-item" title="Fastest Run">
                    <span class="stat-label">Fastest:</span> ${formatDuration(stats.shortestTime)}
                </div>
                <div class="run-stat-item" title="Slowest Run">
                    <span class="stat-label">Slowest:</span> ${formatDuration(stats.longestTime)}
                </div>
                <div class="run-stat-item" title="Average Run">
                    <span class="stat-label">Average:</span> ${formatDuration(stats.averageTime)}
                </div>
                <div class="run-stat-item" title="Total Runs">
                    <span class="stat-label">Total:</span> ${stats.runCount}
                </div>
                <div class="run-stat-item" title="Errors">
                    <span class="stat-label">Errors:</span> ${stats.errorCount}
                </div>
                <div class="run-stat-item" title="Chickens">
                    <span class="stat-label">Chickens:</span> ${stats.runChickens}
                </div>
                <div class="run-stat-item" title="Deaths">
                    <span class="stat-label">Deaths:</span> ${stats.runDeaths}
                </div>
            </div>
        `;
        runStatsGrid.appendChild(runElement);
    }

        runStatsElement.appendChild(runStatsGrid);
    }   


    function calculateRunStats(games) {
        if (!games || games.length === 0) {
            return {};
        }

        const runStats = {};

        games.forEach(game => {
            if (game.Runs && Array.isArray(game.Runs)) {
                game.Runs.forEach(run => {
                    if (!runStats[run.Name]) {
                        runStats[run.Name] = { 
                            shortestTime: Infinity, 
                            longestTime: 0, 
                            totalTime: 0,
                            errorCount: 0, 
                            runCount: 0,
                            runChickens: 0,
                            runDeaths: 0,
                            successfulRunCount: 0,
                            isCurrentRun: false
                        };
                    }

                    // Check if this is the current run
                    if (run.Reason === "") {
                        runStats[run.Name].isCurrentRun = true;
                    }

                    const runTime = new Date(run.FinishedAt) - new Date(run.StartedAt);
                    if (run.FinishedAt !== "0001-01-01T00:00:00Z" && runTime > 0) {
                        runStats[run.Name].runCount++;

                        if (run.Reason === 'ok') {
                            runStats[run.Name].shortestTime = Math.min(runStats[run.Name].shortestTime, runTime);
                            runStats[run.Name].longestTime = Math.max(runStats[run.Name].longestTime, runTime);
                            runStats[run.Name].totalTime += runTime;
                            runStats[run.Name].successfulRunCount++;
                        }
                    }

                    if (run.Reason == 'error') {
                        runStats[run.Name].errorCount++;
                    }

                    if (run.Reason == 'chicken') {
                        runStats[run.Name].runChickens++;
                    }

                    if (run.Reason == 'death') {
                        runStats[run.Name].runDeaths++;
                    }
                });
            }
        });

        // Calculate average time for successful runs
        for (const stats of Object.values(runStats)) {
            if (stats.successfulRunCount > 0) {
                stats.averageTime = stats.totalTime / stats.successfulRunCount;
            } else {
                stats.shortestTime = 0;
                stats.longestTime = 0;
                stats.averageTime = 0;
            }
        }

        return runStats;
    }

    function calculateStats(games) {
        if (!games || games.length === 0) {
            return { totalGames: 0, totalChickens: 0, totalDeaths: 0, totalErrors: 0 };
        }

        return games.reduce((acc, game) => {
            acc.totalGames++;
            if (game.Reason === 'chicken') acc.totalChickens++;
            else if (game.Reason === 'death') acc.totalDeaths++;
            else if (game.Reason === 'error') acc.totalErrors++;
            return acc;
        }, { totalGames: 0, totalChickens: 0, totalDeaths: 0, totalErrors: 0 });
    } 

    function formatDuration(ms) {
        if (!isFinite(ms) || ms < 0) {
            return 'N/A';
        }
        const seconds = Math.floor(ms / 1000);
        const minutes = Math.floor(seconds / 60);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        if (days > 0) return `${days}d ${hours % 24}h`;
        if (hours > 0) return `${hours}h ${minutes % 60}m`;
        if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
        return `${seconds}s`;
    }

    function saveExpandedState() {
        const expandedCards = Array.from(document.querySelectorAll('.character-card.expanded'))
            .map(card => card.id);
        localStorage.setItem('expandedCards', JSON.stringify(expandedCards));
    }

    function restoreExpandedState() {
        const expandedCards = JSON.parse(localStorage.getItem('expandedCards')) || [];
        expandedCards.forEach(cardId => {
            const card = document.getElementById(cardId);
            if (card) {
                card.classList.add('expanded');
                const toggleBtn = card.querySelector('.toggle-details i');
                if (toggleBtn) {
                    toggleBtn.style.transform = 'rotate(180deg)';
                }
            }
        });
    }

    document.addEventListener('DOMContentLoaded', function() {
        fetchInitialData();
        connectWebSocket();
        restoreExpandedState();
    });