document.addEventListener('DOMContentLoaded', () => {
    const sections = document.querySelectorAll('section[id]');
    const menuLinks = document.querySelectorAll('.menu a');

    const observerOptions = {
        root: null,
        rootMargin: '-100px 0px -60% 0px',
        threshold: 0
    };

    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const id = entry.target.getAttribute('id');
                menuLinks.forEach(link => link.classList.remove('active'));
                const activeLink = document.querySelector(`.menu a[href="#${id}"]`);
                if (activeLink) {
                    activeLink.classList.add('active');
                }
            }
        });
    }, observerOptions);

    sections.forEach(section => {
        observer.observe(section);
    });

    const senhaInput = document.getElementById('senha');
    const listaRequisitos = document.getElementById('requisitos');

    // Função que valida a senha e atualiza a checklist — reutilizável
    function validarSenha() {
        const valor = senhaInput.value;

        if (valor.length > 0) {
            listaRequisitos.classList.add('visivel');
        } else {
            listaRequisitos.classList.remove('visivel');
        }

        const checks = {
            'req-maiuscula': /[A-Z]/.test(valor),
            'req-minuscula': /[a-z]/.test(valor),
            'req-numero': /[0-9]/.test(valor),
            'req-especial': /[!@#$%^&*(),.?":{}|<>_\-+=]/.test(valor),
            'req-tamanho': valor.length >= 8
        };

        Object.keys(checks).forEach(id => {
            const item = document.getElementById(id);
            const icone = item.querySelector('.icone');

            if (checks[id]) {
                item.classList.add('ok');
                icone.textContent = '✓';
            } else {
                item.classList.remove('ok');
                icone.textContent = '✕';
            }
        });

        // Retorna true só se TODOS os requisitos foram cumpridos
        return Object.values(checks).every(valor => valor === true);
    }

    if (senhaInput) {
        senhaInput.addEventListener('input', validarSenha);
    }

    const botaoCadastro = document.getElementById('btn-cadastro');
    const emailInput = document.getElementById('email');
    const emailErro = document.getElementById('email-erro');
    const senhaErro = document.getElementById('senha-erro');

    // Domínios de e-mail aceitos (no máximo 4)
    const dominiosValidos = ['gmail.com', 'hotmail.com', 'outlook.com', 'yahoo.com'];

    if (botaoCadastro) {
        botaoCadastro.addEventListener('click', () => {
            const valorSenha = senhaInput.value;

            // Confere os requisitos de senha (sem mexer na visibilidade da checklist)
            const senhaValida =
                /[A-Z]/.test(valorSenha) &&
                /[a-z]/.test(valorSenha) &&
                /[0-9]/.test(valorSenha) &&
                /[!@#$%^&*(),.?":{}|<>_\-+=]/.test(valorSenha) &&
                valorSenha.length >= 8;

            // Mensagem de erro de senha: só avisa se o campo estiver vazio.
            // Se estiver preenchido mas incompleto, a checklist já mostra o que falta.
            if (valorSenha.length === 0) {
                senhaErro.textContent = 'Digite uma senha.';
                senhaErro.classList.add('visivel');
            } else {
                senhaErro.classList.remove('visivel');
            }

            // Validação de e-mail
            const email = emailInput.value.trim();
            let mensagemErro = '';

            if (email.length === 0) {
                mensagemErro = 'Preencha o campo de e-mail.';
            } else if (!email.includes('@')) {
                mensagemErro = 'E-mail inválido: falta o "@".';
            } else {
                const partes = email.split('@');
                const usuario = partes[0];
                const dominio = partes[1];

                if (usuario.length === 0) {
                    mensagemErro = 'E-mail inválido: falta o nome antes do "@".';
                } else if (!dominio || !dominiosValidos.includes(dominio.toLowerCase())) {
                    mensagemErro = `Domínio não aceito. Use: ${dominiosValidos.join(', ')}.`;
                }
            }

            if (mensagemErro) {
                emailErro.textContent = mensagemErro;
                emailErro.classList.add('visivel');
            } else {
                emailErro.classList.remove('visivel');
            }

            // BLOQUEIO: só segue adiante se e-mail E senha estiverem válidos
            if (!senhaValida || mensagemErro) {
                return; // interrompe aqui, não deixa "completar o cadastro"
            }

            console.log('Tudo válido! Aqui entraria a lógica real de cadastro.');
        });
    }
});