let selectedPack = null;

async function register() {
    const username = document.getElementById('regUser').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPass').value;
    const confirm = document.getElementById('regConfirm').value;
    const errorEl = document.getElementById('error');

    if (password !== confirm) {
        errorEl.innerText = "Passwords do not match!";
        return;
    }

    const response = await fetch('/api/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password })
    });

    const data = await response.json();
    if (data.success) {
        alert("Registration successful! Please login.");
        window.location.href = '/login';
    } else {
        errorEl.innerText = data.message;
    }
}

async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const errorEl = document.getElementById('error');

    const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    const data = await response.json();
    if (data.success) {
        window.location.href = '/shop';
    } else {
        errorEl.innerText = data.message;
    }
}

function selectPack(name, price, element) {
    selectedPack = { name, price };
    document.querySelectorAll('.pack-card').forEach(card => card.classList.remove('selected'));
    element.classList.add('selected');
}

async function processPayment() {
    const gameId = document.getElementById('gameId').value;
    if (!gameId || !selectedPack) return alert("Fill all details!");

    const response = await fetch('/api/topup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ gameId, pack: selectedPack })
    });

    if (response.status === 401) return window.location.href = '/login';
    
    const data = await response.json();
    if (data.success) window.location.href = `success.html?id=${data.orderId}`;
}

async function loadOrders() {
    const response = await fetch('/api/orders');
    if (response.status === 401) return window.location.href = '/login';
    
    const orders = await response.json();
    const tbody = document.querySelector('#orderTable tbody');
    tbody.innerHTML = orders.map(o => `
        <tr>
            <td>#${o.id}</td>
            <td>${o.gameId}</td>
            <td>${o.packName}</td>
            <td>$${o.amount}</td>
            <td>${new Date(o.timestamp).toLocaleDateString()}</td>
        </tr>
    `).join('');
}
