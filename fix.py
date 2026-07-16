import re
with open('web/templates/partials/home/intake.html', 'r', encoding='utf-8') as f:
    c = f.read()

for i in range(1, 6):
    if i > 1:
        c = c.replace(f'<div x-show="step === {i}"', f'<div class="col-start-1 row-start-1 w-full" x-show="step === {i}"')
    
    pattern = f'<div class="col-start-1 row-start-1 w-full" x-show="step === {i}"(?:.*?)class="space-y-[45]">'
    new_trans = f'<div class="col-start-1 row-start-1 w-full" x-show="step === {i}" x-cloak x-transition:enter="transition ease-out duration-500 delay-200" x-transition:enter-start="opacity-0 translate-x-4" x-transition:enter-end="opacity-100 translate-x-0" x-transition:leave="transition ease-in duration-200" x-transition:leave-start="opacity-100 translate-x-0" x-transition:leave-end="opacity-0 -translate-x-4" class="space-y-4">'
    c = re.sub(pattern, new_trans, c)

c = re.sub(r':class="\(!contactEmail.*?border"', 'class="btn-submit min-w-[120px]"\n                                :class="submitStatus !== \'idle\' ? \'submitting\' : \'\'"', c, flags=re.DOTALL)
c = c.replace('<span x-show="submitStatus !== \'idle\'">', '<span x-show="submitStatus !== \'idle\' && submitStatus !== \'error\'" class="flex items-center space-x-2">')
c = c.replace('</button>', '<span x-show="submitStatus === \'error\'">{{.T.Get "RetryBtn"}}</span>\n                        </button>')

c = c.replace('text-[10px] mt-1" x-show="emailTouched', 'text-[10px] mt-2 mb-4" x-show="emailTouched')

c = c.replace('.T.Get "CompanyNameHint"', '.T.Get "IntakeStep1Sub"')
c = c.replace('.T.Get "ProjectScopeHint"', '.T.Get "IntakeStep2Sub"')
c = c.replace('.T.Get "DeadlineStepHint"', '.T.Get "IntakeStep3Sub"')
c = c.replace('.T.Get "ContactStepHint"', '.T.Get "IntakeStep5Sub"')
c = c.replace('class="text-xs text-neutral-400"', 'class="text-[11px] text-neutral-400 leading-normal"')

with open('web/templates/partials/home/intake.html', 'w', encoding='utf-8') as f:
    f.write(c)
