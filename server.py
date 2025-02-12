import os
import smtplib
from email.mime.text import MIMEText
from flask import Flask, render_template, request, jsonify
from database import init_db, add_email, remove_email
from dotenv import load_dotenv

load_dotenv()
app = Flask(__name__)
init_db()

def send_confirmation_email(email):
    SMTP_SERVER = os.getenv('SMTP_SERVER')
    SMTP_PORT = int(os.getenv('SMTP_PORT', 465))
    SMTP_USER = os.getenv('SMTP_USER')
    SMTP_PASSWORD = os.getenv('SMTP_PASSWORD')

    # Create the message for the
    unsubscribe_link = f"{request.url_root}unsubscribe?email={email}"
    subject = 'Thanks for subscribing!'
    body = ("Thanks for subscribing!\n\n"
            "We're excited to share our journey with you.\n\n"
            f"If you ever wish to unsubscribe, please click <a href='{unsubscribe_link}'>here</a>."
    )

    msg = MIMEText(body, 'html', 'utf-8')
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

if __name__ == "__main__":
    app.run(debug=True)
