document.addEventListener('DOMContentLoaded', () => {
    const chatForm = document.getElementById('chat-form');
    const messageInput = document.getElementById('message-input');
    const chatMessages = document.getElementById('chat-messages');
    const sendButton = document.getElementById('send-button');

    let isGenerating = false;

    chatForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const message = messageInput.value.trim();
        if (!message || isGenerating) return;

        // Disable input and button while generating
        isGenerating = true;
        messageInput.disabled = true;
        sendButton.disabled = true;

        // Add user message to chat
        appendMessage('user', message);
        messageInput.value = '';

        try {
            const response = await fetch('/api/v1/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    messages: [{
                        role: "user",
                        content: message
                    }],
                    model: "llama2",
                    stream: true
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let assistantMessage = '';

            // Create a placeholder for the assistant's message
            const assistantDiv = document.createElement('div');
            assistantDiv.className = 'message assistant-message';
            chatMessages.appendChild(assistantDiv);

            while (true) {
                const { value, done } = await reader.read();
                if (done) break;

                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');

                for (const line of lines) {
                    if (!line) continue;
                    if (line === 'data: [DONE]') continue;

                    try {
                        const parsed = JSON.parse(line.replace('data: ', ''));
                        if (parsed.choices && parsed.choices[0].delta.content) {
                            assistantMessage += parsed.choices[0].delta.content;
                            assistantDiv.textContent = assistantMessage;
                            chatMessages.scrollTop = chatMessages.scrollHeight;
                        }
                    } catch (e) {
                        console.error('Error parsing chunk:', e);
                    }
                }
            }
        } catch (error) {
            console.error('Error:', error);
            appendMessage('assistant', 'Sorry, there was an error processing your request.');
        } finally {
            // Re-enable input and button
            isGenerating = false;
            messageInput.disabled = false;
            sendButton.disabled = false;
            messageInput.focus();
        }
    });

    function appendMessage(role, content) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${role}-message`;
        messageDiv.textContent = content;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // Handle textarea height
    messageInput.addEventListener('input', () => {
        messageInput.style.height = 'auto';
        messageInput.style.height = messageInput.scrollHeight + 'px';
    });

    // Handle Ctrl+Enter or Cmd+Enter to submit
    messageInput.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            chatForm.dispatchEvent(new Event('submit'));
        }
    });
}); 