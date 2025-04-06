import os
import smtplib
from email.mime.text import MIMEText
from flask import Flask, render_template, request, jsonify
from database import get_connection, init_db, add_email, remove_email
from dotenv import load_dotenv
from collections import namedtuple

load_dotenv()
app = Flask(__name__)
init_db()

def send_confirmation_email(email):
    SMTP_SERVER = os.getenv('SMTP_SERVER')
    SMTP_PORT = int(os.getenv('SMTP_PORT', 465))
    SMTP_USER = os.getenv('SMTP_USER')
    SMTP_PASSWORD = os.getenv('SMTP_PASSWORD')

    unsubscribe_link = f"{request.url_root}unsubscribe?email={email}"

    subject = 'Thanks for subscribing!'

    html_body = render_template(
        'confirmation_email.html',
        unsubscribe_link=unsubscribe_link
    )

    msg = MIMEText(html_body, 'html', 'utf-8')  # Specify HTML
    msg['Subject'] = subject
    msg['From'] = SMTP_USER
    msg['To'] = email

    try:
        server = smtplib.SMTP_SSL(SMTP_SERVER, SMTP_PORT, timeout=10)
        server.login(SMTP_USER, SMTP_PASSWORD)
        server.sendmail(SMTP_USER, email, msg.as_string())
        server.quit()
    except Exception as e:
        print(f"Failed to send email to {email}: {e}")
        
@app.route("/")
def index():
    return render_template("index.html")

@app.route("/subscribe", methods=["POST"])
def subscribe():
    data = request.get_json()
    email = data.get('email')
    if not email:
        return jsonify({"error": "No email provided"}), 400

    if add_email(email):
        send_confirmation_email(email)
        return jsonify({"message": "Email has been added"}), 201
    else:
        return jsonify({"error": "Email already exists"}), 400

@app.route("/unsubscribe", methods=["GET"])
def unsubscribe():
    email = request.args.get("email")
    if not email:
        return "No email specified.", 400

    if remove_email(email):
        return f"The email {email} has been unsubscribed.", 200
    else:
        return f"Email {email} was not found or has already been unsubscribed.", 400

@app.route("/newsletters")
def newsletters():
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute("SELECT id, subject, body, sent_at FROM newsletters ORDER BY sent_at DESC")
    results = cursor.fetchall()
    newsletters = [
        {"id": rec[0], "subject": rec[1], "body": rec[2], "sent_at": rec[3]} 
        for rec in results
    ]
    cursor.close()
    conn.close()
    return render_template("newsletters.html", newsletters=newsletters)

@app.route("/newsletter/<int:newsletter_id>")
def newsletter_detail(newsletter_id):
    conn = get_connection()
    cursor = conn.cursor()
    cursor.execute("SELECT id, subject, body, sent_at FROM newsletters WHERE id = %s", (newsletter_id,))
    record = cursor.fetchone()
    cursor.close()
    conn.close()

    if record is None:
        return "Newsletter not found.", 404

    newsletter = {
        "id": record[0],
        "subject": record[1],
        "body": record[2],
        "sent_at": record[3]
    }
    return render_template("newsletter_detail.html", newsletter=newsletter)

if __name__ == "__main__":
    app.run(debug=True)
