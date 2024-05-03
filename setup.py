from setuptools import setup
setup(
    name = 'glurmo',
    version = '0.1.0',
    packages = ['glurmo'],
    entry_points = {
        'console_scripts': [
            'glurmo = glurmo.__main__:main'
        ]
    })