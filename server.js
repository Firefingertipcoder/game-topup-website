const express = require('express');
const bodyParser = require('body-parser');
const session = require('express-session');
const bcrypt = require('bcrypt');
const db = require('./database');
const path = require('path');

const app = express();
const PORT = 3000;

app.use(bodyParser.json());
app.use(express.static('public'));
app.use(session({
    secret: 'super-secret-gaming-key',
    resave: false,
    saveUninitialized: true
}));

// Middleware to protect routes
const checkAuth = (req, res, next) => {
    if (req.session.user) next();
    else res.status(401).json({ success: false, message: "Please login first!" });
};

// --- AUTH ROUTES ---

// Registration
app.post('/api/register', async (req, res) => {
    const { username, email, password } = req.body;
    try {
        const hashedPassword = await bcrypt.hash(password, 10);
        db.run(`INSERT INTO users (username, email, password) VALUES (?, ?, ?)`, 
            [username, email, hashedPassword], (err) => {
            if (err) return res.status(400).json({ success: false, message: "Username or Email already exists!" });
            res.json({ success: true });
        });
    } catch (e) { res.status(500).json({ success: false, message: "Server error" }); }
});

// Login
app.post('/api/login', (req, res) => {
    const { username, password } = req.body;
    db.get(`SELECT * FROM users WHERE username = ?`, [username], async (err, user) => {
        if (user && await bcrypt.compare(password, user.password)) {
            req.session.user = { id: user.id, username: user.username };
            res.json({ success: true });
        } else {
            res.status(401).json({ success: false, message: "Invalid credentials" });
        }
    });
});

app.get('/api/logout', (req, res) => {
    req.session.destroy();
    res.redirect('/login.html');
});

// --- GAME ROUTES ---

app.post('/api/topup', checkAuth, (req, res) => {
    const { gameId, pack } = req.body;
    const userId = req.session.user.id;

    db.run(`INSERT INTO orders (userId, gameId, packName, amount, status) VALUES (?, ?, ?, ?, ?)`, 
        [userId, gameId, pack.name, pack.price, 'Completed'], function(err) {
        if (err) return res.status(500).json({ success: false, message: "DB Error" });
        res.json({ success: true, orderId: this.lastID });
    });
});

// Get User Order History
app.get('/api/orders', checkAuth, (req, res) => {
    const userId = req.session.user.id;
    db.all(`SELECT * FROM orders WHERE userId = ? ORDER BY timestamp DESC`, [userId], (err, rows) => {
        if (err) return res.status(500).json({ success: false });
        res.json(rows);
    });
});

app.listen(PORT, () => console.log(`🚀 Server live at http://localhost:${PORT}`));