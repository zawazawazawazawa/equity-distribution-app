// Global variables
let studOpponentCount = 0;

// Generate card selector HTML
function createCardSelector(id, placeholder = "Select") {
    const ranks = ['A', 'K', 'Q', 'J', 'T', '9', '8', '7', '6', '5', '4', '3', '2'];
    const suits = [
        { value: 's', symbol: '♠', color: 'black' },
        { value: 'h', symbol: '♥', color: 'red' },
        { value: 'd', symbol: '♦', color: 'red' },
        { value: 'c', symbol: '♣', color: 'black' }
    ];
    
    let html = `<select class="card-select" id="${id}" onchange="validateCards()">`;
    html += `<option value="">${placeholder}</option>`;
    
    for (const rank of ranks) {
        for (const suit of suits) {
            const value = rank + suit.value;
            const display = rank + suit.symbol;
            const color = suit.color === 'red' ? 'style="color: red;"' : '';
            html += `<option value="${value}" ${color}>${display}</option>`;
        }
    }
    
    html += '</select>';
    return html;
}

// Replace all card inputs with selectors on page load
function replaceCardInputsWithSelectors() {
    // PLO cards
    for (let i = 1; i <= 5; i++) {
        replaceInputWithSelector(`ploHand${i}`, 'Hand ' + i);
        replaceInputWithSelector(`ploOpp${i}`, 'Opp ' + i);
    }
    
    for (let i = 1; i <= 5; i++) {
        const labels = ['Flop 1', 'Flop 2', 'Flop 3', 'Turn', 'River'];
        replaceInputWithSelector(`ploBoard${i}`, labels[i-1]);
    }
    
    // Stud cards
    replaceInputWithSelector('studYourDown1', 'Down 1');
    replaceInputWithSelector('studYourDown2', 'Down 2');
    replaceInputWithSelector('studYourDown3', 'Down 3');
    replaceInputWithSelector('studYourUp1', '3rd St');
    replaceInputWithSelector('studYourUp2', '4th St');
    replaceInputWithSelector('studYourUp3', '5th St');
    replaceInputWithSelector('studYourUp4', '6th St');
    
    replaceInputWithSelector('studOppDown1', 'Down 1');
    replaceInputWithSelector('studOppDown2', 'Down 2');
    replaceInputWithSelector('studOppDown3', 'Down 3');
    replaceInputWithSelector('studOppUp1', '3rd St');
    replaceInputWithSelector('studOppUp2', '4th St');
    replaceInputWithSelector('studOppUp3', '5th St');
    replaceInputWithSelector('studOppUp4', '6th St');
}

// Replace a single input with selector
function replaceInputWithSelector(id, placeholder) {
    const input = document.getElementById(id);
    if (input) {
        const selector = createCardSelector(id, placeholder);
        const temp = document.createElement('div');
        temp.innerHTML = selector;
        input.parentNode.replaceChild(temp.firstChild, input);
    }
}

// Validate cards for duplicates
function validateCards() {
    const allSelects = document.querySelectorAll('.card-select');
    const selectedCards = new Map();
    
    // Clear all error states
    allSelects.forEach(select => {
        select.classList.remove('error');
    });
    
    // Check for duplicates
    allSelects.forEach(select => {
        const value = select.value;
        if (value) {
            if (selectedCards.has(value)) {
                // Mark both as errors
                select.classList.add('error');
                selectedCards.get(value).classList.add('error');
            } else {
                selectedCards.set(value, select);
            }
        }
    });
}

// Update interface based on selected game
function updateGameInterface() {
    const gameType = document.getElementById('gameType').value;
    
    // Hide all interfaces
    document.getElementById('ploInterface').style.display = 'none';
    document.getElementById('studInterface').style.display = 'none';
    document.getElementById('commonSettings').style.display = 'none';
    document.getElementById('results').style.display = 'none';
    clearError();
    
    if (!gameType) return;
    
    // Show common settings
    document.getElementById('commonSettings').style.display = 'block';
    
    // Show appropriate interface
    if (gameType === 'plo4' || gameType === 'plo5') {
        document.getElementById('ploInterface').style.display = 'block';
        
        // Show/hide 5th card input for PLO5
        const hand5 = document.getElementById('ploHand5');
        const opp5 = document.getElementById('ploOpp5');
        if (gameType === 'plo5') {
            hand5.style.display = 'inline-block';
            opp5.style.display = 'inline-block';
        } else {
            hand5.style.display = 'none';
            opp5.style.display = 'none';
            hand5.value = '';
            opp5.value = '';
        }
    } else {
        document.getElementById('studInterface').style.display = 'block';
    }
}

// Update PLO opponent input type
function updatePLOOpponentInput() {
    const type = document.getElementById('ploOpponentType').value;
    if (type === 'range') {
        document.getElementById('ploRangeInput').style.display = 'block';
        document.getElementById('ploSpecificInput').style.display = 'none';
    } else {
        document.getElementById('ploRangeInput').style.display = 'none';
        document.getElementById('ploSpecificInput').style.display = 'block';
    }
}

// Update Stud opponent input type
function updateStudOpponentInput() {
    const type = document.getElementById('studOpponentType').value;
    if (type === 'single') {
        document.getElementById('studSingleOpponent').style.display = 'block';
        document.getElementById('studRangeOpponents').style.display = 'none';
    } else {
        document.getElementById('studSingleOpponent').style.display = 'none';
        document.getElementById('studRangeOpponents').style.display = 'block';
        // Add first opponent if none exist
        if (studOpponentCount === 0) {
            addStudOpponent();
        }
    }
}

// Add a new Stud opponent
function addStudOpponent() {
    studOpponentCount++;
    const container = document.getElementById('studOpponentsList');
    
    const opponentDiv = document.createElement('div');
    opponentDiv.className = 'opponent-hand';
    opponentDiv.id = `studOpponent${studOpponentCount}`;
    
    opponentDiv.innerHTML = `
        <h4>Opponent ${studOpponentCount}
            <button type="button" onclick="removeStudOpponent(${studOpponentCount})" class="btn-remove">Remove</button>
        </h4>
        <div class="stud-cards">
            <div class="card-section">
                <span class="card-label">Down Cards:</span>
                <div class="card-inputs">
                    ${createCardSelector(`studOpp${studOpponentCount}Down1`, 'Down 1')}
                    ${createCardSelector(`studOpp${studOpponentCount}Down2`, 'Down 2')}
                    ${createCardSelector(`studOpp${studOpponentCount}Down3`, 'Down 3')}
                </div>
            </div>
            <div class="card-section">
                <span class="card-label">Up Cards:</span>
                <div class="card-inputs">
                    ${createCardSelector(`studOpp${studOpponentCount}Up1`, '3rd St')}
                    ${createCardSelector(`studOpp${studOpponentCount}Up2`, '4th St')}
                    ${createCardSelector(`studOpp${studOpponentCount}Up3`, '5th St')}
                    ${createCardSelector(`studOpp${studOpponentCount}Up4`, '6th St')}
                </div>
            </div>
        </div>
    `;
    
    container.appendChild(opponentDiv);
}

// Remove a Stud opponent
function removeStudOpponent(id) {
    const element = document.getElementById(`studOpponent${id}`);
    if (element) {
        element.remove();
    }
}

// Parse card string (e.g., "As" -> "As")
function parseCard(cardStr) {
    if (!cardStr) return null;
    cardStr = cardStr.trim().replace(/10/, 'T');
    
    // Validate card format
    const validRanks = '23456789TJQKA';
    const validSuits = 'shdc';
    
    if (cardStr.length !== 2) return null;
    
    const rank = cardStr[0].toUpperCase();
    const suit = cardStr[1].toLowerCase();
    
    if (!validRanks.includes(rank) || !validSuits.includes(suit)) {
        return null;
    }
    
    return rank + suit;
}

// Get cards from select elements
function getCards(ids) {
    const cards = [];
    for (const id of ids) {
        const select = document.getElementById(id);
        if (select && select.value) {
            cards.push(select.value);
        }
    }
    return cards;
}

// Check for duplicate cards
function checkDuplicates(allCards) {
    const seen = new Set();
    for (const card of allCards) {
        if (seen.has(card)) {
            return true;
        }
        seen.add(card);
    }
    return false;
}

// Display error message
function showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

// Clear error message
function clearError() {
    const errorDiv = document.getElementById('errorMessage');
    errorDiv.style.display = 'none';
}

// Calculate equity
async function calculateEquity() {
    clearError();
    const gameType = document.getElementById('gameType').value;
    const precision = document.getElementById('precision').value;
    
    if (!gameType) {
        showError('Please select a game type');
        return;
    }
    
    const btn = document.getElementById('calculateBtn');
    btn.disabled = true;
    btn.innerHTML = 'Calculating... <span class="loading"></span>';
    
    try {
        let response;
        
        if (gameType === 'plo4' || gameType === 'plo5') {
            response = await calculatePLOEquity(gameType, precision);
        } else {
            response = await calculateStudEquity(gameType, precision);
        }
        
        if (response.ok) {
            const data = await response.json();
            displayResults(data, gameType);
        } else {
            const error = await response.json();
            showError(error.error || 'Calculation failed');
        }
    } catch (error) {
        showError('Network error: ' + error.message);
    } finally {
        btn.disabled = false;
        btn.innerHTML = 'Calculate Equity';
    }
}

// Calculate PLO equity
async function calculatePLOEquity(gameType, precision) {
    const cardCount = gameType === 'plo5' ? 5 : 4;
    const handIds = [];
    for (let i = 1; i <= cardCount; i++) {
        handIds.push(`ploHand${i}`);
    }
    
    const hand = getCards(handIds);
    if (hand.length !== cardCount) {
        throw new Error(`Please enter all ${cardCount} cards for your hand`);
    }
    
    const boardIds = ['ploBoard1', 'ploBoard2', 'ploBoard3', 'ploBoard4', 'ploBoard5'];
    const board = getCards(boardIds);
    
    if (board.length > 0 && board.length < 3) {
        throw new Error('Board must have at least 3 cards (flop)');
    }
    
    const requestBody = {
        game_type: gameType,
        hand: hand,
        board: board.length > 0 ? board : undefined,
        precision: precision
    };
    
    // Add opponent based on type
    const oppType = document.getElementById('ploOpponentType').value;
    if (oppType === 'range') {
        const rangeStr = document.getElementById('ploRange').value.trim();
        if (rangeStr) {
            requestBody.opponent_range = rangeStr;
        }
    } else {
        const oppIds = [];
        for (let i = 1; i <= cardCount; i++) {
            oppIds.push(`ploOpp${i}`);
        }
        const oppHand = getCards(oppIds);
        if (oppHand.length === cardCount) {
            requestBody.opponent_hand = oppHand;
        }
    }
    
    // Check for duplicates
    const allCards = [...hand, ...board];
    if (requestBody.opponent_hand) {
        allCards.push(...requestBody.opponent_hand);
    }
    if (checkDuplicates(allCards)) {
        throw new Error('Duplicate cards detected');
    }
    
    return fetch('http://localhost:8080/api/v1/equity', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestBody)
    });
}

// Calculate Stud equity
async function calculateStudEquity(gameType, precision) {
    const yourDownIds = ['studYourDown1', 'studYourDown2', 'studYourDown3'];
    const yourUpIds = ['studYourUp1', 'studYourUp2', 'studYourUp3', 'studYourUp4'];
    
    const yourDownCards = getCards(yourDownIds);
    const yourUpCards = getCards(yourUpIds);
    
    if (yourDownCards.length === 0 && yourUpCards.length === 0) {
        throw new Error('Please enter at least some cards for your hand');
    }
    
    const oppType = document.getElementById('studOpponentType').value;
    
    if (oppType === 'single') {
        // Single opponent
        const oppDownIds = ['studOppDown1', 'studOppDown2', 'studOppDown3'];
        const oppUpIds = ['studOppUp1', 'studOppUp2', 'studOppUp3', 'studOppUp4'];
        
        const oppDownCards = getCards(oppDownIds);
        const oppUpCards = getCards(oppUpIds);
        
        const requestBody = {
            your_down_cards: yourDownCards,
            your_up_cards: yourUpCards,
            opponent_down_cards: oppDownCards,
            opponent_up_cards: oppUpCards,
            game_type: gameType,
            precision: precision
        };
        
        // Check for duplicates
        const allCards = [...yourDownCards, ...yourUpCards, ...oppDownCards, ...oppUpCards];
        if (checkDuplicates(allCards)) {
            throw new Error('Duplicate cards detected');
        }
        
        return fetch('http://localhost:8080/api/v1/stud/equity', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });
    } else {
        // Multiple opponents
        const opponents = [];
        const allCards = [...yourDownCards, ...yourUpCards];
        
        for (let i = 1; i <= studOpponentCount; i++) {
            const oppDiv = document.getElementById(`studOpponent${i}`);
            if (!oppDiv) continue;
            
            const downIds = [`studOpp${i}Down1`, `studOpp${i}Down2`, `studOpp${i}Down3`];
            const upIds = [`studOpp${i}Up1`, `studOpp${i}Up2`, `studOpp${i}Up3`, `studOpp${i}Up4`];
            
            const downCards = getCards(downIds);
            const upCards = getCards(upIds);
            
            if (downCards.length > 0 || upCards.length > 0) {
                opponents.push({
                    down_cards: downCards,
                    up_cards: upCards
                });
                allCards.push(...downCards, ...upCards);
            }
        }
        
        if (opponents.length === 0) {
            throw new Error('Please enter at least one opponent');
        }
        
        // Check for duplicates
        if (checkDuplicates(allCards)) {
            throw new Error('Duplicate cards detected');
        }
        
        const requestBody = {
            your_down_cards: yourDownCards,
            your_up_cards: yourUpCards,
            opponent_range: opponents,
            game_type: gameType,
            precision: precision
        };
        
        return fetch('http://localhost:8080/api/v1/stud/range-equity', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });
    }
}

// Display results
function displayResults(data, gameType) {
    const resultsDiv = document.getElementById('results');
    const contentDiv = document.getElementById('resultsContent');
    
    resultsDiv.style.display = 'block';
    
    let html = '';
    
    // Main equity bar
    const equity = data.your_equity || data.equity || 0;
    html += `
        <div class="equity-bar">
            <div class="equity-fill" style="width: ${equity}%">
                <span class="equity-text">${equity.toFixed(2)}%</span>
            </div>
        </div>
    `;
    
    // Details
    html += '<div class="result-details">';
    
    // Game-specific details
    if (gameType.includes('highlow')) {
        // Hi-Lo game details
        if (data.highlow_details) {
            html += '<div class="highlow-details">';
            html += '<h3 style="margin-top: 0; color: #0369a1;">Hi-Lo Split Details</h3>';
            html += `
                <div class="result-item">
                    <span class="result-label">High Pot Win Probability:</span>
                    <span class="result-value">${data.highlow_details.high_equity?.toFixed(2) || '0.00'}%</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Low Pot Win Probability:</span>
                    <span class="result-value">${data.highlow_details.low_equity?.toFixed(2) || '0.00'}%</span>
                </div>
                <div class="result-item">
                    <span class="result-label">Scoop Probability:</span>
                    <span class="result-value">${data.highlow_details.scoop_equity?.toFixed(2) || '0.00'}%</span>
                </div>
            `;
            html += '</div>';
        }
    }
    
    // Common details
    html += `
        <div class="result-item">
            <span class="result-label">Total Iterations:</span>
            <span class="result-value">${data.total_iterations?.toLocaleString() || 'N/A'}</span>
        </div>
    `;
    
    if (data.total_hands) {
        html += `
            <div class="result-item">
                <span class="result-label">Opponent Hands Analyzed:</span>
                <span class="result-value">${data.total_hands}</span>
            </div>
        `;
    }
    
    if (data.calculation_time_ms) {
        html += `
            <div class="result-item">
                <span class="result-label">Calculation Time:</span>
                <span class="result-value">${data.calculation_time_ms.toFixed(2)}ms</span>
            </div>
        `;
    }
    
    // Equity graph for range calculations
    if (data.equity_graph && data.equity_graph.length > 0) {
        html += '<h3 style="margin-top: 20px;">Equity vs Each Opponent:</h3>';
        data.equity_graph.forEach((item, index) => {
            html += `
                <div class="result-item">
                    <span class="result-label">Opponent ${index + 1}:</span>
                    <span class="result-value">${item.equity.toFixed(2)}%</span>
                </div>
            `;
        });
    }
    
    html += '</div>';
    
    contentDiv.innerHTML = html;
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    // Replace all card inputs with selectors
    replaceCardInputsWithSelectors();
});